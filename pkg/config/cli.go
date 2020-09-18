package config

import (
	"errors"
	dyanmic_params "github.com/mostafatalebi/dynamic-params"
)

type ConfigCli struct {
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
	}
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
	cnf.Assertions, _ = cp.GetAsString(FieldAssertBodyString)

	return cnf, nil
}

func (c *ConfigCli) ParseAssertions() (*Config, error) {

}