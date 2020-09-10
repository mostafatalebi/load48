package pkg

import (
	"bytes"
	"fmt"
	"gitlab.com/vdm-shared-packages/logger-go"
	"go.uber.org/atomic"
	"net/http"
	"sync"
	"time"
)

type AdGetLoadTest struct {
	ConcurrentWorkers int
	PerWorker int
	ExecDurationFromHeader bool
	ExecDurationHeaderName string
	CacheUsageHeaderName string
	TargetURL string
	Stats map[string]*AdGetStats
}

type AdGetStats struct {
	Success atomic.Int64
	Failed atomic.Int64
	Total atomic.Int64
	CacheUsed atomic.Int64
	CacheNotUsed atomic.Int64
	TotalDuration time.Duration
	LongestDuration time.Duration
	ShortestDuration time.Duration
	AverageDuration time.Duration
}

func NewAdGetLoadTest(url string) *AdGetLoadTest {
	return &AdGetLoadTest{

		TargetURL: url,
		Stats: make(map[string]*AdGetStats, 0),
	}
}

func (a *AdGetLoadTest) Process() {
	if a.ConcurrentWorkers < 1 || a.PerWorker < 1 {
		logger.Get().Fatal("concurrent & reqCount must be greater than zero")
		return
	}
	wg := &sync.WaitGroup{}
	for i := 0; i < a.ConcurrentWorkers; i++ {
		wg.Add(1)
		fmt.Println("starting worker...")
		go func(workerName string) {
			for j :=  0; j < a.PerWorker; j++ {
				a.SendAdRequest(http.MethodGet, a.TargetURL, workerName)
			}
			wg.Done()
		}(fmt.Sprintf("Worker #%v", i))
	}
	wg.Wait()
}

func (a *AdGetLoadTest) SendAdRequest(method, urlStr, workerName string){
	tn := time.Now()
	if method == http.MethodGet {
		bf := []byte{}
		bd := bytes.NewBuffer(bf)
		cl, _ := http.NewRequest(http.MethodGet, urlStr, bd)
		cl.Header.Set("Origin", "localhost")
		resp, err := http.DefaultClient.Do(cl)
		if err != nil || resp == nil {
			logger.Get().Error("got error", "worker", workerName, "error", err)
		}

		var cacheUsed = false
		if a.CacheUsageHeaderName != "" {
			if resp.Header.Get(a.CacheUsageHeaderName) == "1" {
				cacheUsed = true
			}
		}
		dur := time.Since(tn)
		if a.ExecDurationFromHeader {
			durStr := resp.Header.Get(a.ExecDurationHeaderName)
			if durStr != "" {
				dur, err = time.ParseDuration(durStr)
				if err != nil {
					dur = 0
				}
			}
		}
		if resp.StatusCode == http.StatusOK {
			a.addStat(workerName, "success", dur, cacheUsed)
		} else {
			a.addStat(workerName, "error", dur, cacheUsed)
		}
	}
}

func (a *AdGetLoadTest) addStat(workerName, statType string, dur time.Duration, cacheUsed bool) {
	if _, ok := a.Stats[workerName]; !ok {
		a.Stats[workerName] = &AdGetStats{}
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
	if cacheUsed {
		a.Stats[workerName].CacheUsed.Add(1)
	} else {
		a.Stats[workerName].CacheNotUsed.Add(1)
	}
	a.Stats[workerName].TotalDuration += dur
	a.Stats[workerName].AverageDuration = time.Duration(a.Stats[workerName].TotalDuration.Nanoseconds() / a.Stats[workerName].Total.Load())
}

func (a *AdGetLoadTest) PrintStats(onlyTotal bool) {
	var allStats = &AdGetStats{}
	for k, v := range a.Stats {
		allStats.Total.Add(v.Total.Load())
		allStats.Success.Add(v.Success.Load())
		allStats.Failed.Add(v.Failed.Load())
		allStats.AverageDuration += v.AverageDuration
		allStats.ShortestDuration += v.ShortestDuration
		allStats.LongestDuration += v.LongestDuration
		allStats.CacheUsed.Add(v.CacheUsed.Load())
		allStats.CacheNotUsed.Add(v.CacheNotUsed.Load())
		if onlyTotal == false {
			fmt.Printf("\n=== %v ===", k)
			a.statsPrinter(v)
		}
	}
	a.statsPrinter(allStats)
}

func (a *AdGetLoadTest) statsPrinter(v *AdGetStats) {

	fmt.Printf("\nTotal Number of Requests: %v", v.Total.Load())
	fmt.Printf("\nSuccess: %v", v.Success.Load())
	fmt.Printf("\nFailed: %v", v.Failed.Load())
	fmt.Printf("\nAvergae: %v", v.AverageDuration)
	fmt.Printf("\nLongest: %v", v.LongestDuration)
	fmt.Printf("\nShortest: %v", v.ShortestDuration)
	fmt.Printf("\nWith Cache: %v", v.CacheUsed.Load())
	fmt.Printf("\nWithout Cache: %v", v.CacheNotUsed.Load())
}