package core

import (
	"bytes"
	"fmt"
	dyanmic_params "github.com/mostafatalebi/dynamic-params"
	"github.com/mostafatalebi/loadtest/pkg/stats"
	"log"
	"net"
	"net/http"
	"regexp"
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
	Stats                  *dyanmic_params.DynamicParams
	Lock                  *sync.RWMutex
}

func NewAdGetLoadTest() *LoadTest {
	return &LoadTest{
		Lock: &sync.RWMutex{},
		Url:   "",
		Stats: dyanmic_params.NewDynamicParams(dyanmic_params.SrcNameInternal),
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
		log.Fatalln("concurrentWorkers & perWorker must be greater than zero")
		return
	}
	wg := &sync.WaitGroup{}
	fmt.Printf("starting all workers (%v)...\n", a.ConcurrentWorkers)
	for i := 0; i < a.ConcurrentWorkers; i++ {
		wg.Add(1)
		go func(workerName string) {
			a.AddStat(workerName, stats.NewStatsManager(workerName))
			var bt []byte
			bd := bytes.NewBuffer(bt)
			req, err := http.NewRequest(a.Method, a.Url, bd)
			if err != nil {
				log.Println("error in creating request object: ", err.Error())
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
	if err != nil || resp == nil {
		if ve, ok := err.(net.Error); ok && ve.Timeout() {
			a.GetStat(workerName).IncrTimeout(1)
		} else {
		}
		log.Println("#skip got error:", workerName, err)
		return
	}
	defer resp.Body.Close()

	a.GetStat(workerName).IncrTotal(1)
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
		a.GetStat(workerName).AddShortestDuration(appExecDure)
		a.GetStat(workerName).AddLongestDuration(appExecDure)
		a.GetStat(workerName).AddAverageExecDuration()
	}
	a.GetStat(workerName).IncrCacheUsed(cacheUsed)
	a.GetStat(workerName).AddMainDuration(dur)
	a.GetStat(workerName).AddLongestDuration(dur)
	a.GetStat(workerName).AddShortestDuration(dur)
	a.GetStat(workerName).AddAverageDuration()
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

func (a *LoadTest) PrintPretty(perWorker bool) {
	totalStats := stats.NewStatsManager("total")

	a.Stats.Iterate(func(key string, value interface{}) {
		v, ok := value.(*stats.StatsCollector)
		if !ok {
			return
		}
		if perWorker {
			v.PrintPretty()
		}
		newStats := v.Merge(totalStats)
		newStats.Key = "total"
		totalStats = &newStats
	})

	totalStats.PrintPretty()
}
