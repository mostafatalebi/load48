### LoadTest
LoadTest allows you to send any number of requests, with concurrency settings, to an endpoint and 
provides you with detailed statistics. You can download an executable binary from release section, or
download the source and `make build` it yourself.

---
#### Tables of Contents
1. [Usage](#usage)
    1. [Command Line](#cli)
    2. [Config File](#config-file)
2. [Parameters](#parameters)
2. [Stats](#stats)
3. [Building Source Code](#building-source-code)   
---
##### Usage
In order to use LoadTest, you must pass params to it through command line.

**Cli**
For detailed list of parameters, go to Parameters section.

To send a simple request, go to the directory where `loadtest` executable is, and:
```shell script
loadtest --max-timeout=2 --method=GET --header-Origin=mypc.com --concurrencyt=2 --request-count=10 
--enable-logs=true 
--url="http://127.0.0.1:8081/api/adserver/ad/get?country=spain&domain=youtube.com&format=1"
```
The above command sends 40 requests, with concurrently 2 requests. For example,
if you want to send 40 requests all at the same time, change `--concurrency=40` and 
`--request-count=40`, which
means it sends 40 concurrent requests. You can set any header or use other methods,
sending body with request is not supported yet.

**Config File**
It is not released yet, but it will happen soon.


##### Parameters
Here is the full list of supported parameters.

Make sure to prepend `--` before any param name, if you are going
to use `cli` mode.

`concurrency` **required**
Number of max concurrent requests,

`request-count` **required**
How many workers must be totally sent.

`method` **required**
HTTP method to send the request in, all UPPER CASE.

`url` **required**
The target URL. It can be quoted, if it contains non-regular characters.

`assert-body-string` **optional**
If set, checks to see if response's body's content contains
the given string or not, if yes, then counts that request as success

`max-timeout` **required**
After which the request is considered timed-out. It is the same value
passed to http client, too.

`enable-logs` **optional**
If true, verbose logs are printed.

`exec-duration-header-name` **optional**
This is a nice feature. If your app, in debug mode or test mode, sends a header in its response
which holds the value of internal app execution duration, then you will have a better understanding
of your app. If you implement it, put the name of header for this param and the test will look
for that in the response, too, and if not found, nothing happens. All stats based on this header
are defined with "exec" keyword, apart from general stats. 

For example, if your app sends app exec duration in a header named `App-Duration`, then
you can set the value of this param to `App-Duration`

`cache-usage-header-name` **optional**
If target URL's response headers contain a boolean header than can be used to check if 
the request is being served from cache or not, then set the name of that header to this response

##### Stats
LoadTest produces such stats for test result:
 
If you do not provide `exec-duration-header-name` and `cache-usage-header-name`
they get excluded from the results.

```shell script
======== total ========
--- Total Number of Requests => 400 
--- Shortest Duration => 422.495µs 
--- Shortest App Execution => 29.197µs 
--- Longest Duration => 20.892206ms 
--- Total Success => 400 
--- Average App Execution => 263.95µs 
--- Average Duration => 6.162331ms 
--- Longest App Execution => 220.957µs 

======== Test Info ========
Test Target: 400 // number of initial target set by user
Test Duration: 71.237216ms // total duration of test
Test RAM Usage: 2201KB // RAM usage of whole test

```

##### Building Source Code
If you want to build from the source, download the source code and run:
`make build` it builds and sets the version to last existing tag.