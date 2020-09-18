package config

import (
	"errors"
	"fmt"
	dyanmic_params "github.com/mostafatalebi/dynamic-params"
	"github.com/mostafatalebi/loadtest/pkg/assertions"
	"net/http"
	"regexp"
)

type ConfigCli struct {
	rawArgs []string
}

func NewConfigCli() *ConfigCli {
	return &ConfigCli{}
}

func (c *ConfigCli) LoadConfig(vars ...interface{}) (*Config, error) {
	args := make([]string, 0)
	if vars == nil || len(vars) == 0 {
		return nil, errors.New("os.Args is need as first param")
	} else if v, ok := vars[0].([]string); ok {
		args = v
		c.rawArgs = args
	}
	var err error
	cnf := &Config{}
	var cp = dyanmic_params.NewDynamicParams(dyanmic_params.SrcNameArgs, args)
	cnf.Method, _ = cp.GetAsQuotedString(FieldMethod)
	cnf.Url, _ = cp.GetAsQuotedString(FieldUrl)
	cnInt, _ := cp.GetStringAsInt(FieldConcurrency)
	cnf.Concurrency = int64(cnInt)
	cnInt, _ = cp.GetStringAsInt(FieldNumberOfRequests)
	cnf.NumberOfRequests = int64(cnInt)
	cnf.ExecDurationHeaderName, _ = cp.GetAsString(FieldExecDurationHeaderName)
	cnf.CacheUsageHeaderName, _ = cp.GetAsString(FieldCacheUsageHeaderName)
	cnf.MaxTimeout, _ = cp.GetStringAsInt(FieldMaxTimeout)
	cnf.EnabledLogs, _ = cp.GetStringAsBool(FieldEnableLogs)
	cnf.Assertions, err = c.ParseAssertions()
	if err != nil {
		return nil, errors.New("wrong assertions found")
	} else if cnf.Assertions == nil {
		cnf.Assertions = assertions.NewAssertionManagerWithDefaults(nil)
	}
	cnf.Headers, err = c.ParseHeaders()
	if err != nil {
		return nil, errors.New("wrong headers found")
	}
	cnf.FormBody, _ = cp.GetAsString(FieldFormBody)
	return cnf, nil
}

func (c *ConfigCli) ParseAssertions() (*assertions.AssertionManager, error) {
	var valuesMap = GetMapValuesFromArgs("--assert-", c.rawArgs)
	if valuesMap != nil && len(valuesMap) > 0 {
		var assertionMap = map[string]assertions.Assertion{}
		for k, v := range valuesMap {
			asrt := assertions.NewAssertionFromName(k)
			asrt.SetTest(v)
			assertionMap[k] = asrt
		}
		return assertions.NewAssertionManagerWithDefaults(assertionMap), nil
	}
	return nil, nil
}

func (c *ConfigCli) ParseHeaders() (http.Header, error) {
	var valuesMap = GetMapValuesFromArgs("--header-", c.rawArgs)
	if valuesMap != nil && len(valuesMap) > 0 {
		var headersMap = http.Header{}
		for k, v := range valuesMap {
			headersMap.Set(k, v)
		}
		return headersMap, nil
	}
	return nil, nil
}



// map values in argument might start with a common prefix
// this common prefix is removed from found argument keys
// hence, --header-Origin will become Origin in resulting
// map.
// You can use this function to create a map of values from
// map like keys in the arguments
func GetMapValuesFromArgs(arrPrefix string, args []string) map[string]string {
	valuesMap := map[string]string{}
	rg := regexp.MustCompile(fmt.Sprintf(`%v([a-zA-Z0-9\-]+)\=(.+)`, arrPrefix))
	for _, v := range args {
		if rg.Match([]byte(v)) {
			vals := rg.FindStringSubmatch(v)
			if vals == nil || len(vals) < 2 {
				continue
			}
			valuesMap[vals[1]] = vals[2]
		}
	}
	return valuesMap
}
