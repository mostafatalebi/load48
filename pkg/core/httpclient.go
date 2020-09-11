package core

import (
	"io"
	"log"
	"net/http"
	"sync"
)

var requestOnce = &sync.Once{}
var request *http.Request

func NewHttpClient(method, urlStr string, buff io.Reader) (*http.Request) {
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