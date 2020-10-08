package request

import (
	dyanmic_params "github.com/mostafatalebi/dynamic-params"
	"github.com/mostafatalebi/loadtest/pkg/stats"
	"github.com/mostafatalebi/loadtest/pkg/stats/progress"
	variable "github.com/mostafatalebi/loadtest/pkg/variables"
	"go.uber.org/atomic"
	"sync"
)

const (
	PolicySeq        = "seq"
	PolicyRoundRobin = "round-robin"
	PolicyParallel   = "parallel"
)

type TargetFunc func(variables variable.VariableMap)

var TargetAll = "all target"

type Targeting struct {
	policy                string
	concurrency           int64
	numOfRequests         int64
	Workers               []*RequestWorker
	StatsLock             *sync.RWMutex
	LockConcurrencyStat   *sync.Mutex
	Stats                 *dyanmic_params.DynamicParams
	workersErrors         []error
	eventRequestAttempted chan int8
	eventCCChanged        chan int64
	requestCounter        chan int64
	currentConcurrencyNum atomic.Int64
	progress              *progress.ProgressIndicator
	logFileName           string
}

func NewTargetManager(tp string, c, rc int64) *Targeting {
	t := &Targeting{
		policy:                tp,
		concurrency:           c,
		numOfRequests:         rc,
		LockConcurrencyStat:   &sync.Mutex{},
		StatsLock:             &sync.RWMutex{},
		requestCounter:        make(chan int64, c),
		eventRequestAttempted: make(chan int8),
		eventCCChanged:        make(chan int64),
		progress:              progress.NewProgressIndicator(rc),
	}
	go t.progress.ListenToChannel(t.eventRequestAttempted)
	return t
}

func (t *Targeting) Run() {
	if t.IsSequential() {
		t.SequentialExecution()
	}
}

func (t *Targeting) SequentialExecution() {
	var executionQueue = t.createRecursion(t.Workers, 0)
	wg := sync.WaitGroup{}
	for i := int64(0); i < t.numOfRequests; i++ {
		t.requestCounter <- int64(1)
		wg.Add(1)
		go func() {
			defer func() { <-t.requestCounter }()
			defer wg.Done()
			t.eventRequestAttempted <- 1
			executionQueue(nil)
		}()
	}
	wg.Wait()
	t.MergeTargetsStats()
}

func (t *Targeting)  PrintTargetsStats() {
	for _, v := range t.Workers {
		v.GetStat(v.workerId).PrintPretty(stats.DefaultPresetWithAutoFailedCodes)
	}
}

func (t *Targeting) createRecursion(w []*RequestWorker, index int) TargetFunc {
	if w != nil {
		var reqFunc TargetFunc
		if index == len(w)-1 {
			return nil
		}
		next := t.createRecursion(w, index+1)
		reqFunc = func(variables variable.VariableMap) {
			_, _ = w[index].DoInChain(variables, next)
		}
		return reqFunc
	}
	return nil
}

func (t *Targeting) IsParallel() bool {
	return t.policy == PolicyParallel
}

func (t *Targeting) IsSequential() bool {
	return t.policy == PolicySeq
}

func (t *Targeting) IsRoundRobin() bool {
	return t.policy == PolicyRoundRobin
}

func (t *Targeting) MergeTargetsStats() stats.StatsCollector {
	var totalStats stats.StatsCollector
	for _, ww := range t.Workers {
		ww.Stats.Iterate(func(key string, value interface{}) {
			v, ok := value.(*stats.StatsCollector)
			if !ok {
				return
			}
			newStats := v.Merge(&totalStats)
			v.CalculateAverage()
			newStats.Key = "total"
			totalStats = newStats
		})
	}
	totalStats.CalculateAverage()
	totalStats.CalculateExecAverageDuration()
	return totalStats
}
