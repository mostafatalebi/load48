### Load48 - Load Testing Tool
Load48 is a tool (load-testing tool) for sending managed number of requests to one or more endpoints. It also
provides you with detailed stats of the requests. It has the following features:

- **Yaml Config**
- **Concurrent Requests**
- **Data Sources**
- **Custom Body & Headers**
- **Variables Extraction from JSON Responses**
- **Assertions on Responses**
- **Calculating App Internal Execution with a Custom Header**
- **Multiple Endpoints and Passing Variables Between Them**

#### Installation
Either download an executable binary from releases section
or download the source code and run `sudo make build-linux`
and then you can execute `load48` in your terminal.

#### Sample Configs
From `./examples` directory, you can download a simple or multi-target config file.
Each config file contains comment that might be useful for you to know the fields.

#### Execute Tests
Simply execute the following command in your terminal:
```shell script
load48 --file=path/to/config.yml
```

#### Internals
`load48` works by defining one or more targets in your `.yaml` file. With a "target", we
explicitly mean an endpoint. Each target can have an endpoint url, http method,
 headers, body, assertions
and variable definition from response's body, if it is in JSON format. These variables, if any,
are passed to the next target(s) in row.


#### Yaml Config
Please see examples/ directory for complete config files. But I also put a simple one 
here:
```yaml
main:
  request-count: 100
  concurrency: 10
  targeting-policy: "seq" // the only supported targeting policy now

logs:
  enabled: true
  dir: ./logs

targets:
  login:
    url: http://127.0.0.1:3001/login 
    headers:
      Origin: test.com
      Content-Type: text/html
    max-timeout: 1
    httpMethod: GET
```
Yaml config consists of several sections:

- **main**: Which contains the main parameters of the tests.

- **logs**: contains info about error logging and its directory.

- **data-source**: it is a target which gets executed before the test begins,
and can be used to trigger something on the server or can be used to define
variables from its response (for example, an auth token). Any variable defined
in data-source is usable by all targets.

- **targets**: this is the important section. You can define named targets (for example, login or getUser).
It allows you to define more than one target if you want to have a managed batch of endpoints to be
called. For example, you want to test a scenario which contains getting a product
and then getting its comments. You can define two targets, the first one calls getProduct
and the second one uses info returned by the first one and calls getComments endpoint for that
product.


#### Config Params
`main` `request-count` **int** Number of request per target.

`main` `concurrency` **int**  Number of concurrent requests, this number cannot be greater
than request-count.

`main` `strategy` **string** How to send request: `seq` for sequential, `parallel` for parallel
execution and `round-robin` for a balanced shared of requests for each target.
These values are meaningful only if you have more than one target, otherwise a simple
sequential execution will be used no matter what is the value of strategy. You can
read comments inside sample config files for more explanations.

`logs` `enabled` **bool** Enable error logging.

`logs` `dir` **string** Directory in which error log file is saved. Must have permission,
otherwise the test fails to start.

`target` `url` **string** The url to which request is sent.

`target` `method` **string** HTTP method of the request.

`target` `variables` **map** List of variables (must start with $). Each variable
has `type` and `path`. Path is a dot notation path to search any JSON response.
Variable types are `array`, `map`, `string` and `number`. A sample:
```yaml
# for a json: { "data": { "username" : "bob", "password" : "123456"}}
variables:
    $username: 
        type: string 
        path: data.username
    $password:
        type: string
        path: data.password     
```

`target` `headers` **map** Custom headers. You can use variables defined in previous
targets or in data-source section. You can use variables here, for example:
`Authorization: Bearer $oatuh2Token` and `$oatuh2Token` is a variable defined
either in a data-source or any previous target.

`target` `formBody` **string** A custom body to send with request.

`target` `max-timeout` **int** Number of seconds for a request to be considered timed out.



