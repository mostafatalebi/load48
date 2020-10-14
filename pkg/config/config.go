package config

import (
	"github.com/mostafatalebi/loadtest/pkg/assertions"
	variable "github.com/mostafatalebi/loadtest/pkg/variables"
	"net/http"
)

const (
	FieldExecDurationHeaderName = "exec-duration-header-name"
	FieldCacheUsageHeaderName   = "cache-usage-header-name"
	FieldConcurrency            = "concurrency"
	FieldNumberOfRequests       = "request-count"
	FieldMethod                 = "method"
	FieldUrl                    = "url"
	FieldMaxTimeout             = "max-timeout"
	FieldEnableLogs             = "enable-logs"
	FieldFormBody             = "form-body"
	FieldAssertBodyString       = "assert-body-string"
)



type Config struct {
	Concurrency            int64
	NumberOfRequests       int64
	Method                 string
	TargetName             string
	Url                    string
	MaxTimeout             int
	EnabledLogs            bool
	Assertions             *assertions.AssertionManager
	Headers                http.Header
	FormBody               string
	LogFileDirectory       string
	ExecDurationHeaderName string
	CacheUsageHeaderName   string
	VariablesMap           variable.VariableMap
	Strategy               string `yaml:"-"`
}


type ConfigReader interface {
	LoadConfigs(vars ...interface{}) ([]*Config, error)
	ParseAssertions(data map[string]string) (*assertions.AssertionManager, error)
	ParseHeaders(data map[string]string) (http.Header, error)
	//Validate() []error
}

func NewConfig(configType string) ConfigReader {
	if configType == "yaml" {
		return NewConfigYaml()
	} else if configType == "cli" {
		return NewConfigCli()
	}
	return nil
}


type ConfigStages struct {
	collection map[string]*Config
	current int
}

func NewConfigStages(stages map[string]*Config) {

}

func (cs *ConfigStages) Next() bool {
	return false
}

