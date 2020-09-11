package core

import (
	"bytes"
	"fmt"
	"go.uber.org/atomic"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"
)

var statsMapMx = sync.RWMutex{}

type LoadTest struct {
	ConcurrentWorkers      int
	PerWorker              int
	Method              string
	Url string
	MaxTimeoutSec int
	Headers map[string]string
	ExecDurationFromHeader bool
	ExecDurationHeaderName string
	CacheUsageHeaderName   string
	PerWorkerStats              bool
	Stats                  map[string]*LoadTestStats
}

type LoadTestStats struct {
	Success          atomic.Int64
	Failed           atomic.Int64
	Total            atomic.Int64
	CacheUsed        atomic.Int64
	CacheNotUsed     atomic.Int64
	TotalDuration    time.Duration
	TotalAppExecDuration    time.Duration
	LongestDuration  time.Duration
	LongestAppExecDuration  time.Duration
	ShortestDuration time.Duration
	ShortestAppExecDuration time.Duration
	AverageDuration  time.Duration
	AverageAppExecDuration  time.Duration
}

func NewAdGetLoadTest() *LoadTest {
	return &LoadTest{
		Url: "",
		Stats:     make(map[string]*LoadTestStats, 0),
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
			var bt []byte
			bd := bytes.NewBuffer(bt)
			cl := GetHttpRequestObj(a.Method, a.Url, bd)
			if cl == nil {
				return
			}
			if a.Headers != nil && len(a.Headers) > 0 {
				for hk, hv := range a.Headers {
					cl.Header.Set(hk, hv)
				}
			}
			for j := 0; j < a.PerWorker; j++ {
				a.Send(cl, time.Second * time.Duration(a.MaxTimeoutSec), workerName)
			}
			wg.Done()
		}(fmt.Sprintf("Worker #%v", i))
	}
	wg.Wait()
}

// Prepares a client and sends the actual request, and manages all variables
// needed for stats
func (a *LoadTest) Send(c *http.Request, tout time.Duration, workerName string) {
	tn := time.Now()
	resp, err := GetHttpClient(tout).Do(c)
	if err != nil || resp == nil {
		log.Println("#skip got error:", workerName, err)
		return
	}
	defer resp.Body.Close()

	var cacheUsed = false
	if a.CacheUsageHeaderName != "" {
		if resp.Header.Get(a.CacheUsageHeaderName) == "1" {
			cacheUsed = true
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
	}
	if resp.StatusCode == http.StatusOK {
		a.addStat(workerName, "success", dur, cacheUsed, &appExecDure)
	} else {
		a.addStat(workerName, "error", dur, cacheUsed, &appExecDure)
	}
}

func (a *LoadTest) addStat(workerName, statType string, dur time.Duration, cacheUsed bool, execDuration *time.Duration) {
	statsMapMx.Lock()
	if _, ok := a.Stats[workerName]; !ok {
		a.Stats[workerName] = &LoadTestStats{}
	}
	if statType == "success" {
		a.Stats[workerName].Success.Add(1)
	} else if statType == "error" {
		a.Stats[workerName].Failed.Add(1)
	}
	a.Stats[workerName].Total.Add(1)

	if dur < a.Stats[workerName].ShortestDuration {
		a.Stats[workerName].ShortestDuration = dur
	} else if a.Stats[workerName].ShortestDuration == 0 {
		a.Stats[workerName].ShortestDuration = dur
	}
	if dur > a.Stats[workerName].LongestDuration {
		a.Stats[workerName].LongestDuration = dur
	}
	if execDuration != nil {
		if *execDuration < a.Stats[workerName].ShortestAppExecDuration {
			a.Stats[workerName].ShortestAppExecDuration = *execDuration
		} else if a.Stats[workerName].ShortestAppExecDuration == 0 {
			a.Stats[workerName].ShortestAppExecDuration = *execDuration
		}
		if *execDuration > a.Stats[workerName].LongestAppExecDuration {
			a.Stats[workerName].LongestAppExecDuration += *execDuration
		}
	}
	if cacheUsed {
		a.Stats[workerName].CacheUsed.Add(1)
	} else {
		a.Stats[workerName].CacheNotUsed.Add(1)
	}
	a.Stats[workerName].TotalDuration += dur
	if execDuration != nil {
		a.Stats[workerName].TotalAppExecDuration += *execDuration
	}
	a.Stats[workerName].AverageDuration = time.Duration(a.Stats[workerName].TotalDuration.Nanoseconds() / a.Stats[workerName].Total.Load())
	if execDuration != nil {
		a.Stats[workerName].AverageAppExecDuration = time.Duration(a.Stats[workerName].TotalAppExecDuration.Nanoseconds() / a.Stats[workerName].Total.Load())
	}
	statsMapMx.Unlock()
}

func (a *LoadTest) PrintStats() {
	var allStats = &LoadTestStats{}
	for k, v := range a.Stats {
		allStats.Total.Add(v.Total.Load())
		allStats.Success.Add(v.Success.Load())
		allStats.Failed.Add(v.Failed.Load())
		allStats.AverageDuration += v.AverageDuration
		allStats.ShortestDuration += v.ShortestDuration
		allStats.LongestDuration += v.LongestDuration
		allStats.LongestAppExecDuration += v.LongestAppExecDuration
		allStats.ShortestAppExecDuration += v.ShortestAppExecDuration
		allStats.AverageAppExecDuration += v.AverageAppExecDuration
		allStats.CacheUsed.Add(v.CacheUsed.Load())
		allStats.CacheNotUsed.Add(v.CacheNotUsed.Load())
		if a.PerWorkerStats == true {
			fmt.Printf("\n=== %v ===", k)
			a.statsPrinter(v)
		}
	}
	a.statsPrinter(allStats)
}

func (a *LoadTest) statsPrinter(v *LoadTestStats) {

	fmt.Printf("\nTotal Number of Requests: %v", v.Total.Load())
	fmt.Printf("\nSuccess: %v", v.Success.Load())
	fmt.Printf("\nFailed: %v", v.Failed.Load())
	fmt.Printf("\nAvergae: %v", v.AverageDuration)
	fmt.Printf("\nLongest: %v", v.LongestDuration)
	fmt.Printf("\nShortest: %v", v.ShortestDuration)
	fmt.Printf("\nWith Cache: %v", v.CacheUsed.Load())
	fmt.Printf("\nWithout Cache: %v", v.CacheNotUsed.Load())
	if a.ExecDurationFromHeader {
		fmt.Printf("\nApp Exec Avergae: %v", v.AverageAppExecDuration)
		fmt.Printf("\nApp Exec Longest: %v", v.LongestAppExecDuration)
		fmt.Printf("\nApp Exec Shortest: %v", v.ShortestAppExecDuration)
	}
}


func (a *LoadTest) GetHeadersFromArgs(args []string) map[string]string {
	hds := map[string]string{}
	rg := regexp.MustCompile(`\-\-header-([a-zA-Z0-9\-]+)\=(.+)`)
	for _, v := range args {
		if rg.Match([]byte(v)) {
			vals := rg.FindStringSubmatch(v)
			if vals == nil || len(vals) < 2 {
				continue
			}
			hds[vals[0]] = vals[1]
		}
	}
	if len(hds) > 0 {
		return hds
	}
	return nil
}