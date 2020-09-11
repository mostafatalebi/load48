package main

import (
	"github.com/mostafatalebi/dynamic-params"
	"github.com/mostafatalebi/loadtest/pkg/common"
	"github.com/mostafatalebi/loadtest/pkg/core"
	"os"
)

func main() {
	cp := dyanmic_params.NewDynamicParams(dyanmic_params.SrcNameArgs, os.Args)
	httpMethod, _ := cp.GetAsQuotedString(common.FieldMethod)
	urlVal, _ := cp.GetAsQuotedString(common.FieldUrl)
	workerCount, _ := cp.GetStringAsInt(common.FieldWorkerCount)
	perWorker, _ := cp.GetStringAsInt(common.FieldPerWorker)
	execDebugHeaderName, _ := cp.GetAsString(common.FieldExecDurationHeaderName)
	cacheUsageHeaderName, _ := cp.GetAsString(common.FieldCacheUsageHeaderName)
	perWorkerStats, _ := cp.GetStringAsBool(common.FieldPerWorkerStats)
	lt := core.NewAdGetLoadTest()
	lt.Url = urlVal
	lt.Method = httpMethod
	lt.Headers = lt.GetHeadersFromArgs(os.Args)
	lt.ConcurrentWorkers = workerCount
	lt.PerWorker = perWorker
	lt.PerWorkerStats = perWorkerStats
	if cp.Has(common.FieldExecDurationHeaderName) {
		lt.ExecDurationFromHeader = true
		lt.ExecDurationHeaderName = execDebugHeaderName
	}
	if cp.Has(common.FieldCacheUsageHeaderName) {
		lt.CacheUsageHeaderName = cacheUsageHeaderName
	}
	lt.Process()
	lt.PrintStats()
}
