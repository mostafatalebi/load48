### Load Test Utility
This little program allows you to send a number of requests to an endpoint.

- it supports custom headers
- it allows you to define number of concurrent workers sending requests,
and how many each worker should send
- it gives per worker stats, as well as total stats
- It allows you to define a cache header, and the stats tell you how many
requests have been served by cache how many not
- It allows you to define a custom header, which is used to calculate app exec
duration in the stats (because default duration contains network as well), this 
feature allows you to have both app and total duration stats. Though it requires
you to send your custom header in the API's response. It must contain a value
in duration format, for example: `1ms`


#### Usage
You can download binaries from releases section, or you can build it yourself:
`make build` this command builds the source
code and sets version info of last tag. Use `make buildlatest` to set version info
of last commit.

#### Example Request
The following example uses all possible parameters. For special values such as
URL, don't forget to wrap them in quotation marks to avoid any break.
```shell script
--url=http://example.com/endpoint?token=something
--method GET
--worker-count=1
--per-worker=50
--header-Content-Type=text/html
--header-Origin=http://test.example.com
--header-Authorization=someValue
--exec-duration-header-name=App-Exec-Duration
--cache-usage-header-name=Is-Cache-Used
--per-worker-stats=1
```

#### List of Params

`--url` `string` `required` Target URL to send request to.
`--method` `string` `required` HTTP method of the request
`--worker-count` `int` `required` The number of concurrent request-sending-worker.

`--per-worker` `int` `required` Number of sequential requests each worker sends.

`--header-*` `string` `optional` Any param starting with `--header-` will be treated as a request
header

`--exec-duration-header-name` `string` `optional` You can set you server to send a response header
in debug or test mode which holds the real app duration for that request, and its format
must be in duration format (`1s`, `256ms` etc.). Valid units are Valid time units are "ns", 
"us" (or "µs"), "ms", "s", "m", "h".

`cache-usage-header-name` `string` `optional` A response header which holds a "0" or "1" value
and determines if app has served this request from cache

`--per-worker-stats` `bool` `optional` if set to true, then per worker stats are 
also printed.