package main

import (
	"github.com/mostafatalebi/dynamic-params"
	"github.com/mostafatalebi/loadtest/pkg/common"
	"github.com/mostafatalebi/loadtest/pkg/core"
	"github.com/mostafatalebi/loadtest/pkg/logger"
	"github.com/mostafatalebi/loadtest/pkg/stats"
	"net/http"
	"os"
	"strings"
)

var Version = ""

func main() {
	CheckCommandEntry()
	cp := dyanmic_params.NewDynamicParams(dyanmic_params.SrcNameArgs, os.Args)
	httpMethod, _ := cp.GetAsQuotedString(common.FieldMethod)
	urlVal, _ := cp.GetAsQuotedString(common.FieldUrl)
	workerCount, _ := cp.GetStringAsInt(common.FieldWorkerCount)
	perWorker, _ := cp.GetStringAsInt(common.FieldPerWorker)
	execDebugHeaderName, _ := cp.GetAsString(common.FieldExecDurationHeaderName)
	cacheUsageHeaderName, _ := cp.GetAsString(common.FieldCacheUsageHeaderName)
	maxTimeout, _ := cp.GetStringAsInt(common.FieldMaxTimeout)
	enableLogs, _ := cp.GetStringAsBool(common.FieldEnableLogs)
	lt := core.NewAdGetLoadTest()
	lt.EnableLogs = enableLogs
	logger.LogEnabled = enableLogs
	lt.Url = urlVal
	if httpMethod == "" {
		httpMethod = http.MethodGet
	}
	lt.Method = strings.ToUpper(httpMethod)
	lt.Headers = lt.GetHeadersFromArgs(os.Args)
	lt.ConcurrentWorkers = workerCount
	lt.PerWorker = perWorker
	lt.MaxTimeoutSec = maxTimeout
	if cp.Has(common.FieldExecDurationHeaderName) {
		lt.ExecDurationFromHeader = true
		lt.ExecDurationHeaderName = execDebugHeaderName
	}
	if cp.Has(common.FieldCacheUsageHeaderName) {
		lt.CacheUsageHeaderName = cacheUsageHeaderName
	}
	lt.Process()
	st := lt.MergeAll()
	st.PrintPretty(stats.DefaultPresetWithAutoFailedCodes)
}

func CheckCommandEntry() {
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		PrintHelp()
		os.Exit(0)
		return
	} else if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		PrintVersion()
		os.Exit(0)
		return
	}
}