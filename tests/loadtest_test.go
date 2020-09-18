package tests

import (
	"github.com/mostafatalebi/loadtest/pkg/logger"
	"net"
	"net/http"
	"os"
	"testing"
)

var listenAddrPort = "13756"

func TestMain(m *testing.M) {
	ls, err := net.Listen("tcp", ":"+listenAddrPort)
	if err != nil {
		panic("failed to run the test, cannot create test http server: "+err.Error())
	}
	logger.Info("running http server", "success")

	r := http.NewServeMux()
	r.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Test-Timeout") != "" {
			writer.WriteHeader(http.StatusGatewayTimeout)
			return
		} else if request.Header.Get("Test-Ok") != "" {
			writer.Write([]byte("hey!"))
			return
		} else if request.Header.Get("Test-Failed") != "" {
			writer.WriteHeader(500)
			return
		}
	})

	go func() {
		if err := http.Serve(ls, r); err != nil {
			panic(err)
		}
	}()
	os.Exit(m.Run())
}

//func TestLoadTestOnMockServer(t *testing.T) {
//	lt := loadtest.NewLoadTest()
//	lt.MaxConcurrentRequests = 100
//	lt.NumberOfRequests = 1
//	lt.Url = "http://127.0.0.1:"+listenAddrPort+"/test"
//	lt.Method = http.MethodGet
//	lt.MaxTimeoutSec = 1
//	lt.Headers = &http.Header{}
//	lt.Headers.Set("Test-Timeout", "1")
//	lt.Process()
//	statsAll := lt.MergeAll()
//	assert.Equal(t, int64(100), statsAll.GetTimeout())
//	statsAll.PrintPretty(stats.DefaultPresetWithAutoFailedCodes)
//}