package main

import (
	"fmt"
	"strings"
)

func PrintHelp() {
	PrintAuthorInformation()
	fmt.Println(`--url string required Target URL to send request to.
	--method string required HTTP method of the request
	--worker-count int required The number of concurrent request-sending-worker.

	--per-worker int required Number of sequential requests each worker sends.

	--header-* string optional Any param starting with --header- will be treated as a request
	header

	--exec-duration-header-name string optional You can set you server to send a response header
	in debug or test mode which holds the real app duration for that request, and its format
	must be in duration format (1s, 256ms etc.). Valid units are Valid time units are "ns",
		"us" (or "µs"), "ms", "s", "m", "h".

	cache-usage-header-name string optional A response header which holds a "0" or "1" value
	and determines if app has served this request from cache

	--per-worker-stats bool optional if set to true, then per worker stats are
	also printed.`)
}

func PrintVersion() {
	PrintAuthorInformation()
	fmt.Printf("Version: %v\n", UnderstandVersion(Version))
}

func UnderstandVersion(v string) string {
	if strings.Contains(v, "/") {
		spl := strings.Split(v, "/")
		if len(spl) > 0 {
			return spl[len(spl)-1]
		}
	} else {
		return v
	}
	return ""
}

func PrintAuthorInformation() {
	fmt.Println("Author: Mostafa Talebi")
	fmt.Println("Email: most.talebi@gmail.com")
}