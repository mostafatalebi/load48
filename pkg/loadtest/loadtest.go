package loadtest

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gojektech/valkyrie"
	dyanmic_params "github.com/mostafatalebi/dynamic-params"
	"github.com/mostafatalebi/loadtest/pkg/assertions"
	"github.com/mostafatalebi/loadtest/pkg/config"
	"github.com/mostafatalebi/loadtest/pkg/logger"
	print2 "github.com/mostafatalebi/loadtest/pkg/print"
	"github.com/mostafatalebi/loadtest/pkg/stats"
	"go.uber.org/atomic"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

var statsMapMx = sync.RWMutex{}

type LoadTest struct {
	Config *config.Config
	MaxConcurrentRequests     int64
	currentConcurrentRequests atomic.Int64
	Stats                  *dyanmic_params.DynamicParams
	Lock                   *sync.RWMutex
	testStartTime          time.Time
	requestChan			   chan int64
}

func NewLoadTest(cnf *config.Config) *LoadTest {
	l := &LoadTest{
		Config: cnf,
		Stats:                     dyanmic_params.NewDynamicParams(dyanmic_params.SrcNameInternal, &sync.RWMutex{}),
		Lock:                      &sync.RWMutex{},
		testStartTime:             time.Time{},
		requestChan:               nil,
	}
	l.currentConcurrentRequests.Add(0)
	return l
}

func (ld *LoadTest) publishRequestsToChannel() {
	if ld.requestChan == nil {
		ld.requestChan = make(chan int64, ld.Config.Concurrency)
	}
	go func() {
		for i := int64(0); i < ld.Config.NumberOfRequests; i++ {
			ld.requestChan <- int64(1)
		}
		close(ld.requestChan)
	}()
}

func (ld *LoadTest) AddStat(name string, s *stats.StatsCollector) {
	ld.Stats.Add(name, s)
}
func (ld *LoadTest) GetStat(name string) *stats.StatsCollector {
	ld.Lock.Lock()
	defer ld.Lock.Unlock()
	s := ld.Stats.Get(name)
	if s == nil {
		return nil
	}
	if v, ok := s.(*stats.StatsCollector); ok {
		return v
	}
	return nil
}
func (ld *LoadTest) Process() error {
	if ld.Config.Concurrency < 1 || ld.Config.NumberOfRequests < 1 {
		logger.Fatal("incorrect params", "concurrent & request-count param must be greater than zero")
		return errors.New("incorrect params")
	} else if ld.Config.NumberOfRequests < ld.Config.Concurrency {
		logger.Fatal("incorrect params", "concurrent cannot be greater than request-count")
		return errors.New("incorrect params")
	}
	ld.publishRequestsToChannel()
	ld.testStartTime = time.Now()
	wg := &sync.WaitGroup{}
	logger.Info("Test Status", fmt.Sprintf("starting...", ld.Config.Concurrency))
	var bt = []byte(ld.Config.FormBody)
	bd := bytes.NewBuffer(bt)
	req, err := http.NewRequest(ld.Config.Method, ld.Config.Url, bd)
	wg.Add(1)
	go func() {
		ld.AddStat("default", stats.NewStatsManager("default"))
		ld.GetStat("default").IncrSuccess(0)
		for _ = range ld.requestChan {
			if err != nil {
				logger.Error("creating request object failed", err.Error())
				return
			}
			req.Header = ld.Config.Headers
			ld.Send(req, time.Second*time.Duration(ld.Config.MaxTimeout), "default")
		}
		wg.Done()
	}()
	wg.Wait()
	return nil
}

// Prepares a client and sends the actual request, and manages all variables
// needed for stats
func (ld *LoadTest) Send(req *http.Request, tout time.Duration, profileName string) {
	defer ld.currentConcurrentRequests.Add(-1)
	tn := time.Now()
	resp, err := GetHttpClient(tout).Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	ld.GetStat(profileName).IncrTotalSent(1)
	err = ld.UnderstandResponse(profileName, resp, err)
	if err != nil {
		logger.Error("request failed", err.Error())
		return
	} else if resp == nil {
		logger.Error("request failed", "no error and no response")
		return
	}

	{
		// assertions on response
		if ld.Config.Assertions.Exists(assertions.AssertBodyString) {
			btData, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Error("failed to read body of response", err)
				ld.GetStat(profileName).IncrOtherErrors(1)
				return
			}
			_ = ld.Config.Assertions.Get(assertions.AssertBodyString).SetInput(btData)
		}
		_ = ld.Config.Assertions.Get(assertions.AssertStatusIsOk).SetTest(resp.StatusCode)
		if err := ld.Config.Assertions.ChainRunner(assertions.AssertStatusIsOk, assertions.AssertBodyString); err == nil {
			ld.GetStat(profileName).IncrSuccess(1)
		} else {
			ld.GetStat(profileName).IncrFailed(resp.StatusCode, 1)
		}
	}

	var cacheUsed = int64(0)
	if ld.Config.CacheUsageHeaderName != "" {
		if resp.Header.Get(ld.Config.CacheUsageHeaderName) == "1" {
			cacheUsed = int64(1)
		}
	}
	dur := time.Since(tn)
	var appExecDure time.Duration
	if ld.Config.ExecDurationHeaderName != "" {
		durStr := resp.Header.Get(ld.Config.ExecDurationHeaderName)
		if durStr != "" {
			appExecDure, err = time.ParseDuration(durStr)
			if err != nil {
				appExecDure = 0
			}
		}
		ld.GetStat(profileName).AddExecDuration(appExecDure)
		ld.GetStat(profileName).AddExecShortestDuration(appExecDure)
		ld.GetStat(profileName).AddExecLongestDuration(appExecDure)
	}
	ld.GetStat(profileName).IncrCacheUsed(cacheUsed)
	ld.GetStat(profileName).AddMainDuration(dur)
	ld.GetStat(profileName).AddLongestDuration(dur)
	ld.GetStat(profileName).AddShortestDuration(dur)
	print2.ProgressByPercent(ld.Config.NumberOfRequests, ld.GetStat("default").GetTotal())
}

func (ld *LoadTest) UnderstandResponse(profileName string, resp *http.Response, err interface{}) error {
	if err != nil || resp == nil {
		if ve, ok := err.(net.Error); ok && ve.Timeout() {
			ld.GetStat(profileName).IncrTimeout(1)
			logger.Error("request timeout", "["+profileName+"]"+ve.Error())
		} else if ve, ok := err.(*valkyrie.MultiError); ok {
			errStr := ve.Error()
			if err := ve.HasError(); strings.Contains(errStr, "context deadline exceeded") {
				ld.GetStat(profileName).IncrTimeout(1)
				return errors.New("context timeout => [" + profileName + "]" + err.Error())
			} else if err := ve.HasError(); strings.Contains(err.Error(), "connect: connection refused") {
				ld.GetStat(profileName).IncrConnRefused(1)
				return errors.New("connection refused => [" + profileName + "]" + err.Error())
			} else {
				ld.GetStat(profileName).IncrOtherErrors(1)
				return errors.New("other errors => [" + profileName + "]" + err.Error())
			}
		} else {
			errStr := ""
			if v, ok := err.(error); ok {
				errStr = v.Error()
			}
			ld.GetStat(profileName).IncrFailed(500, 1)
			return errors.New("other errors => [" + profileName + "]" + errStr)
		}
	} else if resp.StatusCode == 504 {
		ld.GetStat(profileName).IncrTimeout(1)
		return errors.New("server timeout => [" + profileName + "]")
	}
	return nil
}


func (ld *LoadTest) MergeAll() stats.StatsCollector {
	var totalStats stats.StatsCollector

	ld.Stats.Iterate(func(key string, value interface{}) {
		v, ok := value.(*stats.StatsCollector)
		if !ok {
			return
		}
		newStats := v.Merge(&totalStats)
		newStats.Key = "total"
		totalStats = newStats
	})
	totalStats.CalculateAverage()
	totalStats.CalculateExecAverageDuration()
	return totalStats
}

func (ld *LoadTest) PrintGeneralInfo() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Println("\n======== Test Info ========")
	fmt.Printf("Test Target: %v\n", ld.Config.NumberOfRequests)
	fmt.Printf("Test Duration: %v\n", time.Since(ld.testStartTime))
	fmt.Printf("Test RAM Usage: %vKB\n\n", memStats.Alloc/1024)
}
