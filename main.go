package main

import (
	"github.com/mostafatalebi/dynamic-params"
	"github.com/mostafatalebi/loadtest/pkg"
	"os"
)

func main() {
	cp := dyanmic_params.NewDynamicParams(dyanmic_params.SrcNameArgs, os.Args)
	httpMethod, _ := cp.GetAsString("method")
	urlVal, _ := cp.GetAsString("url")
	workerCount, _ := cp.GetStringAsInt("worker-count")
	perWorker, _ := cp.GetStringAsInt("per-worker")
	execDebugHeaderName, _ := cp.GetAsString("exec-debug-header-name")
	cacheUsageHeaderName, _ := cp.GetAsString("cache-usage-header-name")
	lt := pkg.NewAdGetLoadTest()
	lt.Url = urlVal
	lt.Method = httpMethod
	lt.Headers = lt.GetHeadersFromArgs(os.Args)
	lt.ConcurrentWorkers = workerCount
	lt.PerWorker = perWorker
	if cp.Has("exec-debug-header-name") {
		lt.ExecDurationFromHeader = true
		lt.ExecDurationHeaderName = execDebugHeaderName
	}
	if cp.Has("cache-usage-header-name") {
		lt.CacheUsageHeaderName = cacheUsageHeaderName
	}
	lt.Process()
	lt.PrintStats(true)
}
