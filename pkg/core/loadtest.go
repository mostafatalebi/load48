package core

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gojektech/valkyrie"
	dyanmic_params "github.com/mostafatalebi/dynamic-params"
	"github.com/mostafatalebi/loadtest/pkg/logger"
	"github.com/mostafatalebi/loadtest/pkg/stats"
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
	ConcurrentWorkers      int
	PerWorker              int
	Method                 string
	Url                    string
	MaxTimeoutSec          int
	Headers                *http.Header
	ExecDurationFromHeader bool
	ExecDurationHeaderName string
	CacheUsageHeaderName   string
	PerWorkerStats         bool
	EnableLogs			bool
	Stats                  *dyanmic_params.DynamicParams
	Lock                  *sync.RWMutex
	testStartTime time.Time
}

func NewAdGetLoadTest() *LoadTest {
	return &LoadTest{
		Lock: &sync.RWMutex{},
		Url:   "",
		Stats: dyanmic_params.NewDynamicParams(dyanmic_params.SrcNameInternal, &sync.RWMutex{}),
	}
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
func (a *LoadTest) Process() {
	if a.ConcurrentWorkers < 1 || a.PerWorker < 1 {
		logger.Fatal("incorrect params", "concurrentWorkers & perWorker must be greater than zero")
		return
	}
	a.testStartTime = time.Now()
	wg := &sync.WaitGroup{}
	logger.Info("Test Status", fmt.Sprintf("starting workers(%v)", a.ConcurrentWorkers))
	for i := 0; i < a.ConcurrentWorkers; i++ {
		wg.Add(1)
		go func(workerName string) {
			a.AddStat(workerName, stats.NewStatsManager(workerName))
			var bt []byte
			bd := bytes.NewBuffer(bt)
			req, err := http.NewRequest(a.Method, a.Url, bd)
			if err != nil {
				logger.Error("creating request object failed", err.Error())
				return
			}
			req.Header = *a.Headers
			for j := 0; j < a.PerWorker; j++ {
				a.Send(req, time.Second*time.Duration(a.MaxTimeoutSec), workerName)
			}
			wg.Done()
		}(fmt.Sprintf("Worker #%v", i))
	}
	wg.Wait()
}

// Prepares a client and sends the actual request, and manages all variables
// needed for stats
func (a *LoadTest) Send(req *http.Request, tout time.Duration, workerName string) {
	tn := time.Now()
	resp, err := GetHttpClient(tout).Do(req)
	defer resp.Body.Close()
	a.GetStat(workerName).IncrTotal(1)
	err = a.UnderstandResponse(workerName, resp, err)

	if err != nil {
		logger.Error("request failed", err.Error())
	}

	if resp.StatusCode == 200 {
		a.GetStat(workerName).IncrSuccess(1)
	} else {
		a.GetStat(workerName).IncrFailed(resp.StatusCode, 1)
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
		a.GetStat(workerName).AddExecDuration(appExecDure)
		a.GetStat(workerName).AddExecShortestDuration(appExecDure)
		a.GetStat(workerName).AddExecLongestDuration(appExecDure)
		a.GetStat(workerName).AddExecAverageDuration()
	}
	a.GetStat(workerName).IncrCacheUsed(cacheUsed)
	a.GetStat(workerName).AddMainDuration(dur)
	a.GetStat(workerName).AddLongestDuration(dur)
	a.GetStat(workerName).AddShortestDuration(dur)
	a.GetStat(workerName).AddAverageDuration()
}

func (a *LoadTest) UnderstandResponse(workerName string, resp *http.Response, err interface{}) error {
	if err != nil || resp == nil {
		if ve, ok := err.(net.Error); ok && ve.Timeout() {
			a.GetStat(workerName).IncrTimeout(1)
			logger.Error("request timeout", "["+workerName+"]" + ve.Error())
		} else if ve, ok := err.(*valkyrie.MultiError); ok  {
			errStr := ve.Error()
			if err := ve.HasError(); strings.Contains(errStr, "context deadline exceeded") {
				a.GetStat(workerName).IncrTimeout(1)
				return errors.New("context timeout => ["+workerName+"]" + err.Error())
			} else if err := ve.HasError(); strings.Contains(err.Error(), "connect: connection refused") {
				a.GetStat(workerName).IncrConnRefused(1)
				return errors.New("connection refused => ["+workerName+"]" + err.Error())
			} else {
				a.GetStat(workerName).IncrOtherErrors(1)
				return errors.New("other errors => ["+workerName+"]" + err.Error())
			}
		} else {
			errStr := ""
			if v, ok := err.(error); ok {
				errStr = v.Error()
			}
			a.GetStat(workerName).IncrFailed(500, 1)
			return errors.New("other errors => ["+workerName+"]" + errStr)
		}
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

	return totalStats
}


func (a *LoadTest) PrintGeneralInfo() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Println("\n======== Test Info ========")
	fmt.Printf("Test Duration: %v\n", time.Since(a.testStartTime))
	fmt.Printf("Test RAM Usage: %vKB\n\n", memStats.Alloc/1024)
}
