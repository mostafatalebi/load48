package core

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gojektech/valkyrie"
	dyanmic_params "github.com/mostafatalebi/dynamic-params"
	"github.com/mostafatalebi/loadtest/pkg/config"
	"github.com/mostafatalebi/loadtest/pkg/logger"
	"github.com/mostafatalebi/loadtest/pkg/stats"
	"go.uber.org/atomic"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
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
	NumberOfRequests          int64
	Method                    string
	AssertBodyString          string
	Url                       string
	MaxTimeoutSec             int
	Headers                   *http.Header
	ExecDurationFromHeader    bool
	ExecDurationHeaderName    string
	CacheUsageHeaderName      string
	PerWorkerStats            bool
	EnableLogs                bool
	Stats                  *dyanmic_params.DynamicParams
	Lock                   *sync.RWMutex
	testStartTime          time.Time
	requestChan			   chan int64
}

func NewLoadTest(cnf *config.Config) *LoadTest {
	l := &LoadTest{
		MaxConcurrentRequests:     cnf.Concurrency,
		NumberOfRequests:          cnf.NumberOfRequests,
		Method:                    cnf.Method,
		AssertBodyString:          cnf.AssertBodyString,
		Url:                       cnf.Url,
		MaxTimeoutSec:             cnf.MaxTimeout,
		ExecDurationHeaderName:    cnf.ExecDurationHeaderName,
		CacheUsageHeaderName:      cnf.CacheUsageHeaderName,
		EnableLogs:                cnf.EnabledLogs,
		Stats:                     dyanmic_params.NewDynamicParams(dyanmic_params.SrcNameInternal, &sync.RWMutex{}),
		Lock:                      &sync.RWMutex{},
		testStartTime:             time.Time{},
		requestChan:               nil,
	}
	l.currentConcurrentRequests.Add(0)
	return l
}

func (a *LoadTest) publishRequestsToChannel() {
	if a.requestChan == nil {
		a.requestChan = make(chan int64, a.MaxConcurrentRequests)
	}
	go func() {
		for i := int64(0); i < a.NumberOfRequests; i++ {
			a.requestChan <- int64(1)
		}
		close(a.requestChan)
	}()
}

func (a *LoadTest) AddStat(name string, s *stats.StatsCollector) {
	a.Stats.Add(name, s)
}
func (a *LoadTest) GetStat(name string) *stats.StatsCollector {
	a.Lock.Lock()
	defer a.Lock.Unlock()
	s := a.Stats.Get(name)
	if s == nil {
		return nil
	}
	if v, ok := s.(*stats.StatsCollector); ok {
		return v
	}
	return nil
}
func (a *LoadTest) Process() error {
	if a.MaxConcurrentRequests < 1 || a.NumberOfRequests < 1 {
		logger.Fatal("incorrect params", "concurrent & request-count param must be greater than zero")
		return errors.New("incorrect params")
	} else if a.NumberOfRequests < a.MaxConcurrentRequests {
		logger.Fatal("incorrect params", "concurrent cannot be greater than request-count")
		return errors.New("incorrect params")
	}
	a.publishRequestsToChannel()
	a.testStartTime = time.Now()
	wg := &sync.WaitGroup{}
	logger.Info("Test Status", fmt.Sprintf("starting workers(%v)", a.MaxConcurrentRequests))
	var bt []byte
	bd := bytes.NewBuffer(bt)
	req, err := http.NewRequest(a.Method, a.Url, bd)
	wg.Add(1)
	go func() {
		a.AddStat("default", stats.NewStatsManager("default"))
		a.GetStat("default").IncrSuccess(0)
		for _ = range a.requestChan {
			if err != nil {
				logger.Error("creating request object failed", err.Error())
				return
			}
			req.Header = *a.Headers
			a.Send(req, time.Second*time.Duration(a.MaxTimeoutSec), "default")
		}
		wg.Done()
	}()
	wg.Wait()
	return nil
}

// Prepares a client and sends the actual request, and manages all variables
// needed for stats
func (a *LoadTest) Send(req *http.Request, tout time.Duration, profileName string) {
	defer a.currentConcurrentRequests.Add(-1)
	tn := time.Now()
	resp, err := GetHttpClient(tout).Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	a.GetStat(profileName).IncrTotalSent(1)
	err = a.UnderstandResponse(profileName, resp, err)
	if err != nil {
		logger.Error("request failed", err.Error())
		return
	} else if resp == nil {
		logger.Error("request failed", "no error and no response")
		return
	}
	if resp.StatusCode == 200 {
		if a.AssertBodyString != "" {
			btdata := []byte{}
			btdata, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				bdstr := string(btdata)
				if strings.Contains(bdstr, a.AssertBodyString) {
					a.GetStat(profileName).IncrSuccess(1)
				}
			}

		} else {
			a.GetStat(profileName).IncrSuccess(1)
		}
	} else {
		a.GetStat(profileName).IncrFailed(resp.StatusCode, 1)
	}
	var cacheUsed = int64(0)
	if a.CacheUsageHeaderName != "" {
		if resp.Header.Get(a.CacheUsageHeaderName) == "1" {
			cacheUsed = int64(1)
		}
	}
	dur := time.Since(tn)
	var appExecDure time.Duration
	if a.ExecDurationFromHeader {
		durStr := resp.Header.Get(a.ExecDurationHeaderName)
		if durStr != "" {
			appExecDure, err = time.ParseDuration(durStr)
			if err != nil {
				appExecDure = 0
			}
		}
		a.GetStat(profileName).AddExecDuration(appExecDure)
		a.GetStat(profileName).AddExecShortestDuration(appExecDure)
		a.GetStat(profileName).AddExecLongestDuration(appExecDure)
	}
	a.GetStat(profileName).IncrCacheUsed(cacheUsed)
	a.GetStat(profileName).AddMainDuration(dur)
	a.GetStat(profileName).AddLongestDuration(dur)
	a.GetStat(profileName).AddShortestDuration(dur)
}

func (a *LoadTest) UnderstandResponse(profileName string, resp *http.Response, err interface{}) error {
	if err != nil || resp == nil {
		if ve, ok := err.(net.Error); ok && ve.Timeout() {
			a.GetStat(profileName).IncrTimeout(1)
			logger.Error("request timeout", "["+profileName+"]"+ve.Error())
		} else if ve, ok := err.(*valkyrie.MultiError); ok {
			errStr := ve.Error()
			if err := ve.HasError(); strings.Contains(errStr, "context deadline exceeded") {
				a.GetStat(profileName).IncrTimeout(1)
				return errors.New("context timeout => [" + profileName + "]" + err.Error())
			} else if err := ve.HasError(); strings.Contains(err.Error(), "connect: connection refused") {
				a.GetStat(profileName).IncrConnRefused(1)
				return errors.New("connection refused => [" + profileName + "]" + err.Error())
			} else {
				a.GetStat(profileName).IncrOtherErrors(1)
				return errors.New("other errors => [" + profileName + "]" + err.Error())
			}
		} else {
			errStr := ""
			if v, ok := err.(error); ok {
				errStr = v.Error()
			}
			a.GetStat(profileName).IncrFailed(500, 1)
			return errors.New("other errors => [" + profileName + "]" + errStr)
		}
	} else if resp.StatusCode == 504 {
		a.GetStat(profileName).IncrTimeout(1)
		return errors.New("server timeout => [" + profileName + "]")
	}
	return nil
}

func (a *LoadTest) GetHeadersFromArgs(args []string) *http.Header {
	hds := &http.Header{}
	rg := regexp.MustCompile(`\-\-header-([a-zA-Z0-9\-]+)\=(.+)`)
	for _, v := range args {
		if rg.Match([]byte(v)) {
			vals := rg.FindStringSubmatch(v)
			if vals == nil || len(vals) < 2 {
				continue
			}
			hds.Add(vals[1], vals[2])
		}
	}
	return hds
}

func (a *LoadTest) MergeAll() stats.StatsCollector {
	var totalStats stats.StatsCollector

	a.Stats.Iterate(func(key string, value interface{}) {
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

func (a *LoadTest) PrintGeneralInfo() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Println("\n======== Test Info ========")
	fmt.Printf("Test Target: %v\n", a.NumberOfRequests)
	fmt.Printf("Test Duration: %v\n", time.Since(a.testStartTime))
	fmt.Printf("Test RAM Usage: %vKB\n\n", memStats.Alloc/1024)
}
