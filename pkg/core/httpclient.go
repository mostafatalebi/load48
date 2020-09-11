package core

import (
	"io"
	"log"
	"net/http"
	"sync"
)

var requestOnce = &sync.Once{}
var clientOnce = &sync.Once{}
var request *http.Request
var client *http.Client

func GetHttpRequestObj(method, urlStr string, buff io.Reader) *http.Request {
	requestOnce.Do(func() {
		if request == nil {
			cl, err := http.NewRequest(method, urlStr, buff)
			if err == nil {
				request = cl
			} else {
				log.Fatal("creating client failed")
			}
		}
	})
	return request
}

func GetHttpClient() *http.Client {
	clientOnce.Do(func() {
		if client == nil {
			client = http.DefaultClient
		}
	})
	return client
}
