package loadtest

import (
	"fmt"
	"github.com/mostafatalebi/loadtest/pkg/config"
	"github.com/mostafatalebi/loadtest/pkg/request"
	"github.com/mostafatalebi/loadtest/pkg/stats"
	"runtime"
	"sync"
	"time"
)

var statsMapMx = sync.RWMutex{}

type LoadTest struct {
	testStartTime time.Time
	workers     []*request.RequestWorker
	workersErrors     []error
	requestChan chan int64
}


func NewLoadTest(cnf *config.Config) *LoadTest {
	l := &LoadTest{
		workers: make([]*request.RequestWorker, 0),
	}
	l.workers = append(l.workers, request.NewRequestWorker(cnf))

	return l
}

func (ld *LoadTest) StartWorkers() {
	ld.testStartTime = time.Now()
	if len(ld.workers) == 0 {
		panic("no worker has beenfound to start")
	}
	wg := &sync.WaitGroup{}
	for _, v := range ld.workers {
		wg.Add(1)
		go func () {
			err := v.Do()
			if err != nil {
				ld.workersErrors = append(ld.workersErrors, err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func (ld *LoadTest) PrintWorkersStats() {
	for _, v := range ld.workers {
		v.GetStat(request.DefaultStatsContainer).PrintPretty(stats.DefaultPresetWithAutoFailedCodes)
	}
}
func (ld *LoadTest) PrintGeneralInfo() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Println("\n======== Test Info ========")
	numOfRequest := int64(0)
	for _, w := range ld.workers {
		numOfRequest += w.Config.NumberOfRequests
	}
	fmt.Printf("Test Target: %v\n", numOfRequest)
	fmt.Printf("Test Duration: %v\n", time.Since(ld.testStartTime))
	fmt.Printf("Test RAM Usage: %vKB\n\n", memStats.Alloc/1024)
}
