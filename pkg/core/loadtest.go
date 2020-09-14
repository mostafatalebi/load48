package core

import (
	"bytes"
	"fmt"
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
	Stats                  map[string]*stats.StatsCollector
}

func NewAdGetLoadTest() *LoadTest {
	return &LoadTest{
		Url:   "",
		Stats: make(map[string]*stats.StatsCollector, 0),
	}
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
			a.Stats[workerName] = stats.NewStatsManager(workerName)
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
			a.Stats[workerName].IncrTimeout(1)
		} else {
		}
		log.Println("#skip got error:", workerName, err)
		return
	}
	defer resp.Body.Close()

	a.Stats[workerName].IncrTotal(1)
	if resp.StatusCode == 200 {
		a.Stats[workerName].IncrSuccess(1)
	} else {
		a.Stats[workerName].IncrFailed(resp.StatusCode, 1)
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
		a.Stats[workerName].AddExecDuration(appExecDure)
		a.Stats[workerName].AddShortestDuration(appExecDure)
		a.Stats[workerName].AddLongestDuration(appExecDure)
		a.Stats[workerName].AddAverageExecDuration()
	}
	a.Stats[workerName].IncrCacheUsed(cacheUsed)
	a.Stats[workerName].AddMainDuration(dur)
	a.Stats[workerName].AddLongestDuration(dur)
	a.Stats[workerName].AddShortestDuration(dur)
	a.Stats[workerName].AddAverageDuration()
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

	for _, v := range a.Stats {
		if perWorker {
			v.PrintPretty()
		}
		newStats := v.Merge(totalStats)
		newStats.Key = "total"
		totalStats = &newStats
	}

	totalStats.PrintPretty()
}
