package core

import (
	"github.com/gojektech/heimdall/v6/httpclient"
	"sync"
	"time"
)

var requestOnce = &sync.Once{}
var clientOnce = &sync.Once{}
var client *httpclient.Client

func GetHttpClient(tout time.Duration) *httpclient.Client {
	clientOnce.Do(func() {
		if client == nil {
			client = httpclient.NewClient(httpclient.WithHTTPTimeout(tout))
		}
	})
	return client
}
