package config

import (
	"github.com/mostafatalebi/loadtest/pkg/assertions"
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
	Concurrency int64 `yaml:"concurrency"`
	NumberOfRequests int64 `yaml:"request-count"`
	Method string `yaml:"method"`
	Url string `yaml:"url"`
	MaxTimeout int `yaml:"max-timeout"`
	EnabledLogs bool `yaml:"enable-logs"`
	Assertions *assertions.AssertionManager `yaml:"assertions"`
	Headers http.Header `yaml:"headers"`
	FormBody string `yaml:"form-body"`
	LogFileDirectory string `yaml:"log-dir"`
	ExecDurationHeaderName string `yaml:"exec-duration-header-name"`
	CacheUsageHeaderName   string `yaml:"cache-usage-header-name"`
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

