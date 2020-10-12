package request

import (
	"github.com/mostafatalebi/loadtest/pkg/curr"
	"github.com/mostafatalebi/loadtest/pkg/logger"
	"github.com/mostafatalebi/loadtest/pkg/stats"
	"github.com/mostafatalebi/loadtest/pkg/stats/progress"
	variable "github.com/mostafatalebi/loadtest/pkg/variables"
	"go.uber.org/atomic"
	"sync"
	"time"
)

const (
	PolicySeq        = "seq"
	PolicyRoundRobin = "round-robin"
	PolicyParallel   = "parallel"

	ExecWorker = "w"
	ExecDataSource = "ds"
)

type TargetFunc func(variables variable.VariableMap)

var TargetAll = "all target"

type Targeting struct {
	policy                string
	concurrency           int64
	numOfRequests         int64
	DataSources           []*RequestWorker
	Workers               []*RequestWorker
	StatsLock             *sync.RWMutex
	LockConcurrencyStat   *sync.Mutex
	StatsTotal            *stats.StatsCollector
	workersErrors         []error
	eventRequestAttempted chan int8
	eventCCChanged        chan int64
	requestCounter        chan int64
	currentConcurrencyNum atomic.Int64
	progress              *progress.ProgressIndicator
	logFileName           string
	Variables 			  variable.VariableMap
}

func NewTargetManager(tp string, cc, rc int64) *Targeting {
	t := &Targeting{
		policy:                tp,
		concurrency:           cc,
		numOfRequests:         rc,
		LockConcurrencyStat:   &sync.Mutex{},
		StatsLock:             &sync.RWMutex{},
		requestCounter:        make(chan int64, cc),
		eventRequestAttempted: make(chan int8),
		eventCCChanged:        make(chan int64),
		progress:              progress.NewProgressIndicator(rc),
	}
	go t.progress.ListenToChannel(t.eventRequestAttempted)
	return t
}


// Run accepts an execType which tells it to execute which batch of workers
// because a Run() may mean running actual target workers, or data-sources.
// Run() for data-sources collects stats but yet, it does not do anything to
// them.
// Run() for target workers currently only supports sequential execution
// type, though it has a cascading (recursive) way of executing sibling
// targets.
func (t *Targeting) Run(execType string) {
	if execType == ExecWorker {
		logger.InfoOut("running targets...", "")
		if t.IsSequential() {
			t.SequentialExecution(t.Workers)
		}
	} else if execType == ExecDataSource {
		if t.DataSources != nil {
			logger.InfoOut("running data-source(s)...", "")
			t.SequentialExecutionOfDataSources(t.DataSources)
		}
	} else {
		logger.InfoOut("nothing run, no exec type specified", "")
	}
}

// Creates a reverse recursion from the given list of workers.
// So, the last element in the array becomes the leaf (and hence
// it gets executed last). Each target is responsible to execute
// its next target (hence recursion is important), inherit
// all previous variables from either data-source or previous targets
// and combine them with its own variables (if any defined), and pass
// them to the next target in row.
func (t *Targeting) SequentialExecution(batch []*RequestWorker) {
	var executionQueue = t.createRecursion(batch, 0)
	wg := sync.WaitGroup{}
	for i := int64(0); i < t.numOfRequests; i++ {
		t.requestCounter <- int64(1)
		wg.Add(1)
		go func() {
			defer func() { <-t.requestCounter }()
			defer wg.Done()
			t.eventRequestAttempted <- 1
			executionQueue(t.Variables)
		}()
	}
	wg.Wait()
	t.StatsTotal = t.MergeTargetsStats()
}

// @todo not implemented
func (t *Targeting) ParallelExecution(batch []*RequestWorker) {

}

// This is the same as SequentialExecution(), but is aimed toward
// data-sources and does not change any global stat or does not
// signal any global event.
func (t *Targeting) SequentialExecutionOfDataSources(batch []*RequestWorker) {
	var executionQueue = t.createRecursion(batch, 0)
	if t.DataSources[0].RefreshConfig.RefreshType == "ms" {
		if t.DataSources[0].RefreshConfig.Count < 1 {
			executionQueue(t.Variables)
		} else {
			wt := curr.NewWait(time.Duration(t.DataSources[0].RefreshConfig.Count)*time.Millisecond, 0, 0)
			for wt.Waiting() {
				executionQueue(t.Variables)
			}
		}
	} else if t.DataSources[0].RefreshConfig.RefreshType == "sec" {
		if t.DataSources[0].RefreshConfig.Count < 1 {
			executionQueue(t.Variables)
		} else {
			wt := curr.NewWait(time.Duration(t.DataSources[0].RefreshConfig.Count)*time.Second, 0, 0)
			for wt.Waiting() {
				executionQueue(t.Variables)
			}
		}
	}
}

func (t *Targeting) PrintTargetsStats() {
	for _, v := range t.Workers {
		v.GetStat(v.workerId).PrintPretty(stats.DefaultPresetWithAutoFailedCodes)
	}
	if t.StatsTotal != nil {
		t.StatsTotal.PrintPretty(stats.DefaultPresetWithAutoFailedCodes)
	}
}

// creates a list of functions to be called recursively, this is used
// to nest dependent targets into each other, starting from the leaf target
// (on which no other target is dependent).
// The way it works allows us to transfer defined variables (if any) to next
// target easily. Also, it helps us to maintain max concurrency and stats
// accurate.
func (t *Targeting) createRecursion(w []*RequestWorker, index int) TargetFunc {
	if w != nil {
		var reqFunc, next TargetFunc
		if index == len(w)-1 {
			next = nil
		} else {
			next = t.createRecursion(w, index+1)
		}
		reqFunc = func(vars variable.VariableMap) {
			vars, _ = w[index].DoInChain(vars, next)
			if vars != nil {
				t.Variables = variable.Merge(t.Variables, vars)
			}
		}
		return reqFunc
	}
	return nil
}

// @todo Not Implemented
func (t *Targeting) IsParallel() bool {
	return t.policy == PolicyParallel
}


func (t *Targeting) IsSequential() bool {
	return t.policy == PolicySeq
}

// @todo Not Implemented
func (t *Targeting) IsRoundRobin() bool {
	return t.policy == PolicyRoundRobin
}

func (t *Targeting) MergeTargetsStats() *stats.StatsCollector {
	var totalStats stats.StatsCollector
	for _, ww := range t.Workers {
		ws := ww.Stats.Get(ww.workerId)
		var wsv *stats.StatsCollector
		if ws != nil {
			wsv, _ = ws.(*stats.StatsCollector)
		}
		if wsv == nil {
			return nil
		}
		newStats := wsv.Merge(&totalStats)
		wsv.CalculateAverage()
		newStats.Key = "total"
		totalStats = newStats
	}
	totalStats.CalculateAverage()
	totalStats.CalculateExecAverageDuration()
	return &totalStats
}
