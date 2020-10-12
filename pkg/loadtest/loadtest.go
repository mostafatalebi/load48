package loadtest

import (
	"fmt"
	"github.com/mostafatalebi/loadtest/pkg/config"
	"github.com/mostafatalebi/loadtest/pkg/logger"
	"github.com/mostafatalebi/loadtest/pkg/request"
	"github.com/mostafatalebi/loadtest/pkg/stats"
	"github.com/rs/xid"
	"os"
	"runtime"
	"sync"
	"time"
)

var statsMapMx = sync.RWMutex{}

type LoadTest struct {
	testStartTime time.Time
	workers     []*request.RequestWorker
	workersErrors     []error
	targeting *request.Targeting
}

// each config means a new worker
// for targeting policy, because it is a global config,
// we use first config's targeting policy. This is because
// all configs must have the same targeting policy
func NewLoadTest(configs ...*config.Config) *LoadTest {
	var sessionName = xid.New().String()
	if len(configs) == 0 {
		panic("at least one config must be specified")
	}
	l := &LoadTest{
		workers: make([]*request.RequestWorker, 0),
		targeting: request.NewTargetManager(configs[0].TargetingPolicy, configs[0].Concurrency, configs[0].NumberOfRequests),
	}

	if configs[0].EnabledLogs != true {
		fmt.Println("logs are disabled")
		logger.LogEnabled = false
	} else {
		var logFileName = fmt.Sprintf("loadtest-%v-%v-%v", time.Now().Year(), time.Now().Month(), time.Now().Day()) + sessionName + ".log"
		err := logger.Initialize(logger.LogModeFile,  configs[0].LogFileDirectory+logFileName)
		if err != nil {
			panic(err)
		}
	}

	i := 0
	for _, cc := range configs {
		// @todo it is better to put zero-initializer inside a new function and name the func as InitializeWorker()
		w := request.NewRequestWorker(cc, fmt.Sprintf("%v%v", cc.TargetName, i))
		sm := stats.NewStatsManager(cc.TargetName)
		sm.IncrSuccess(0)
		w.AddStat(fmt.Sprintf("%v%v", cc.TargetName, i), sm)
		l.targeting.Workers = append(l.targeting.Workers, w)
		i++
	}
	return l
}

func (ld *LoadTest) ApplyDataSources(dataSources ...*config.Config) {
	if ld.targeting != nil {
		i := 0
		for _, cc := range dataSources {
			w := request.NewRequestWorker(cc, fmt.Sprintf("data-source:%v%v", cc.TargetName, i))
			sm := stats.NewStatsManager(cc.TargetName)
			sm.IncrSuccess(0)
			w.AddStat(fmt.Sprintf("%v%v", cc.TargetName, i), sm)
			ld.targeting.DataSources = append(ld.targeting.Workers, w)
			i++
		}
	}
}

func (ld *LoadTest) StartWorkers() {
	ld.testStartTime = time.Now()
	if len(ld.targeting.Workers) == 0 {
		fmt.Println("no worker has been found to start")
		os.Exit(1)
	}
	ld.targeting.Run(request.ExecWorker)
}

func (ld *LoadTest) PrintWorkersStats() {
	ld.targeting.PrintTargetsStats()
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
