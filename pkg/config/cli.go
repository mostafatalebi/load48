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

func (c *ConfigCli) LoadConfigs(vars ...interface{}) ([]*Config, error) {
	fmt.Println("Warning: load48 is executed using cli params instead of yaml, it is strongly recommended to use .yaml file instead of command-line params")
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
	if cnInt == 0 {
		return nil, errors.New("[cli] --concurrency cannot be zero")
	}
	cnf.Concurrency = int64(cnInt)
	cnInt, _ = cp.GetStringAsInt(FieldNumberOfRequests)
	if cnInt == 0 {
		return nil, errors.New("[cli] --request-count cannot be zero")
	}
	cnf.NumberOfRequests = int64(cnInt)
	cnf.ExecDurationHeaderName, _ = cp.GetAsString(FieldExecDurationHeaderName)
	cnf.CacheUsageHeaderName, _ = cp.GetAsString(FieldCacheUsageHeaderName)
	cnf.MaxTimeout, _ = cp.GetStringAsInt(FieldMaxTimeout)
	cnf.EnabledLogs, _ = cp.GetStringAsBool(FieldEnableLogs)
	var assertionsMap = GetMapValuesFromArgs("--assert-", c.rawArgs)
	cnf.Assertions, err = c.ParseAssertions(assertionsMap)
	if err != nil {
		return nil, errors.New("wrong assertions found")
	} else if cnf.Assertions == nil {
		cnf.Assertions = assertions.NewAssertionManagerWithDefaults(nil)
	}
	var headersMap = GetMapValuesFromArgs("--header-", c.rawArgs)
	cnf.Headers, err = c.ParseHeaders(headersMap)
	if err != nil {
		return nil, errors.New("wrong headers found")
	}
	cnf.FormBody, _ = cp.GetAsString(FieldFormBody)
	return []*Config{cnf}, nil
}

func (c *ConfigCli) ParseAssertions(valuesMap map[string]string) (*assertions.AssertionManager, error) {
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

func (c *ConfigCli) ParseHeaders(valuesMap map[string]string) (http.Header, error) {

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
