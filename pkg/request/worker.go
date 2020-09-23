package request

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
	"github.com/rs/xid"
	"go.uber.org/atomic"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const DefaultStatsContainer = "default"

// Request worker is responsible to manage sending request
// with all its config to a specific endpoint
type RequestWorker struct {
	Config                 *config.Config
	SessionName            string
	MaxConcurrentRequests  int64
	concurrencyMaxAchieved atomic.Int64
	Stats                  *dyanmic_params.DynamicParams
	Lock                   *sync.RWMutex
	LockConcurrencyStat		   *sync.Mutex
	eventCCChanged		   chan int64
	testStartTime          time.Time
	requestChan            chan int64
	currentConcurrencyNum  atomic.Int64
}

func NewRequestWorker(cnf *config.Config) *RequestWorker {
	var sessionName = xid.New().String()
	r := &RequestWorker{
		Config:        cnf,
		SessionName:   sessionName,
		Stats:         dyanmic_params.NewDynamicParams(dyanmic_params.SrcNameInternal, &sync.RWMutex{}),
		Lock:          &sync.RWMutex{},
		LockConcurrencyStat: &sync.Mutex{},
		eventCCChanged: make(chan int64),
		testStartTime: time.Time{},
		requestChan:   nil,
	}

	if cnf.EnabledLogs != true {
		logger.LogEnabled = false
	} else {
		r.Config.LogFileDirectory = logger.DefaultDirectory
		var logFileName = fmt.Sprintf("loadtest-%v-%v-%v", time.Now().Year(), time.Now().Month(), time.Now().Day()) + sessionName + ".log"
		err := logger.Initialize(logger.LogModeFile, r.Config.LogFileDirectory+logFileName)
		if err != nil {
			panic(err)
		}
	}
	r.requestChan = make(chan int64, r.Config.Concurrency)
	go r.CalculateMaxConcurrency()
	return r
}

func (r *RequestWorker) Do() error {
	if r.Config.Concurrency < 1 || r.Config.NumberOfRequests < 1 {
		logger.Fatal("incorrect params", "concurrent & request-count param must be greater than zero")
		return errors.New("incorrect params")
	} else if r.Config.NumberOfRequests < r.Config.Concurrency {
		logger.Fatal("incorrect params", "concurrent cannot be greater than request-count")
		return errors.New("incorrect params")
	}
	//r.publishRequestsToChannel()
	r.testStartTime = time.Now()
	wg := &sync.WaitGroup{}
	logger.InfoOut("Test Status", fmt.Sprintf("session start: %v", r.SessionName))
	var bt = []byte(r.Config.FormBody)
	bd := bytes.NewBuffer(bt)
	req, err := http.NewRequest(r.Config.Method, r.Config.Url, bd)
	j := 1
	//for _ = range r.requestChan {
	r.AddStat(DefaultStatsContainer, stats.NewStatsManager(DefaultStatsContainer))
	for i := int64(0); i < r.Config.NumberOfRequests; i++ {
		r.requestChan <- int64(1)
		wg.Add(1)
		go func() {
			defer func() { <-r.requestChan }()
			defer wg.Done()
			defer r.UpdateConcurrentReqNum(-1)
			r.UpdateConcurrentReqNum(1)
			r.GetStat(DefaultStatsContainer).IncrSuccess(0)
			if err != nil {
				logger.Error("creating request object failed", err.Error())
				return
			}
			req.Header = r.Config.Headers
			r.sendRequest(req, time.Second*time.Duration(r.Config.MaxTimeout), DefaultStatsContainer)
		}()
		j++
	}
	wg.Wait()
	r.MergeAll()
	return nil
}

func (r *RequestWorker) sendRequest(req *http.Request, tout time.Duration, profileName string) {
	tn := time.Now()
	resp, err := GetHttpClient(tout).Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	err = r.HandleResponse(profileName, resp, err)

	if err != nil {
		logger.Error("request failed", err.Error())
		return
	} else if resp == nil {
		logger.Error("request failed", "no error and no response")
		return
	}

	{
		// assertions on response
		if r.Config.Assertions.Exists(assertions.AssertBodyString) {
			btData, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Error("failed to read body of response", err)
				r.GetStat(profileName).IncrOtherErrors(1)
				return
			}
			_ = r.Config.Assertions.Get(assertions.AssertBodyString).SetInput(btData)
		}
		_ = r.Config.Assertions.Get(assertions.AssertStatusIsOk).SetTest(resp.StatusCode)
		if err := r.Config.Assertions.ChainRunner(assertions.AssertStatusIsOk, assertions.AssertBodyString); err == nil {
			r.GetStat(profileName).IncrSuccess(1)
		} else {
			r.GetStat(profileName).IncrFailed(resp.StatusCode, 1)
		}
	}

	var cacheUsed = int64(0)
	if r.Config.CacheUsageHeaderName != "" {
		if resp.Header.Get(r.Config.CacheUsageHeaderName) == "1" {
			cacheUsed = int64(1)
		}
	}
	dur := time.Since(tn)
	var appExecDure time.Duration
	if r.Config.ExecDurationHeaderName != "" {
		durStr := resp.Header.Get(r.Config.ExecDurationHeaderName)
		if durStr != "" {
			appExecDure, err = time.ParseDuration(durStr)
			if err != nil {
				appExecDure = 0
			}
		}
		r.GetStat(profileName).AddExecDuration(appExecDure)
		r.GetStat(profileName).AddExecShortestDuration(appExecDure)
		r.GetStat(profileName).AddExecLongestDuration(appExecDure)
	}
	r.GetStat(profileName).IncrCacheUsed(cacheUsed)
	r.GetStat(profileName).AddMainDuration(dur)
	r.GetStat(profileName).AddLongestDuration(dur)
	r.GetStat(profileName).AddShortestDuration(dur)
	print2.ProgressByPercent(r.Config.NumberOfRequests, r.GetStat(DefaultStatsContainer).GetTotal())
}

func (r *RequestWorker) HandleResponse(profileName string, resp *http.Response, err interface{}) error {
	if err != nil || resp == nil {
		if ve, ok := err.(net.Error); ok && ve.Timeout() {
			r.GetStat(profileName).IncrTimeout(1)
			logger.Error("request timeout", "["+profileName+"]"+ve.Error())
		} else if ve, ok := err.(net.Error); ok && !ve.Timeout() {
			r.GetStat(profileName).IncrOtherErrors(1)
			logger.Error("request timeout", "["+profileName+"]"+ve.Error())
		} else if ve, ok := err.(*valkyrie.MultiError); ok {
			errStr := ve.Error()
			if err := ve.HasError(); strings.Contains(errStr, "context deadline exceeded") {
				r.GetStat(profileName).IncrTimeout(1)
				r.GetStat(profileName).IncrTotalSent(1)
				return errors.New("context timeout => [" + profileName + "]" + err.Error())
			} else if err := ve.HasError(); strings.Contains(err.Error(), "connect: connection refused") {
				r.GetStat(profileName).IncrConnRefused(1)
				r.GetStat(profileName).IncrTotalSent(1)
				return errors.New("connection refused => [" + profileName + "]" + err.Error())
			} else {
				r.GetStat(profileName).IncrOtherErrors(1)
				r.GetStat(profileName).IncrTotalSent(1)
				return errors.New("other errors => [" + profileName + "]" + err.Error())
			}
		} else {
			errStr := ""
			if v, ok := err.(error); ok {
				errStr = v.Error()
			}
			r.GetStat(profileName).IncrFailed(500, 1)
			r.GetStat(profileName).IncrTotalSent(1)
			return errors.New("other errors => [" + profileName + "]" + errStr)
		}
	} else if resp.StatusCode == 504 {
		r.GetStat(profileName).IncrTimeout(1)
		return errors.New("server timeout => [" + profileName + "]")
	}
	return nil
}

func (r *RequestWorker) AddStat(name string, s *stats.StatsCollector) {
	r.Stats.Add(name, s)
}

func (r *RequestWorker) GetStat(name string) *stats.StatsCollector {
	r.Lock.Lock()
	defer r.Lock.Unlock()
	s := r.Stats.Get(name)
	if s == nil {
		return nil
	}
	if v, ok := s.(*stats.StatsCollector); ok {
		return v
	}
	return nil
}

// the value is a signed +1 or -1, and the worker on the other end
// uses this value to calculate max concurrency achieved
func (r *RequestWorker) UpdateConcurrentReqNum(val int8) {
	r.LockConcurrencyStat.Lock()
	defer r.LockConcurrencyStat.Unlock()
	if val != -1 && val != 1 {
		return
	} else if val == 1 {
		r.currentConcurrencyNum.Add(int64(val))
	} else if val == -1 {
		r.currentConcurrencyNum.Add(int64(val))
	}
	r.eventCCChanged <- r.currentConcurrencyNum.Load()
}

func (r *RequestWorker) CalculateMaxConcurrency() {
	for sig := range r.eventCCChanged {
		r.GetStat(DefaultStatsContainer).UpdateMaxConcurrencyAchieved(sig)
	}
}
func (r *RequestWorker) MergeAll() stats.StatsCollector {
	var totalStats stats.StatsCollector

	r.Stats.Iterate(func(key string, value interface{}) {
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
