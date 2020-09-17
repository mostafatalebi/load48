package core

import (
	"net/http"
	"sync"
	"time"
)

var requestOnce = &sync.Once{}
var clientOnce = &sync.Once{}
var client *http.Client

func GetHttpClient(timeout time.Duration) *http.Client {
	clientOnce.Do(func() {
		if client == nil {
			tr := &http.Transport{
				MaxIdleConnsPerHost: 1024,
				MaxIdleConns: 1024,
			}
			client = &http.Client{}
			client.Timeout = timeout
			client.Transport = tr
		}
	})
	return client
}
