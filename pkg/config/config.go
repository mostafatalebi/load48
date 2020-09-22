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
	Concurrency int64
	NumberOfRequests int64
	Method string
	Url string
	MaxTimeout int
	EnabledLogs bool
	Assertions *assertions.AssertionManager
	Headers http.Header
	FormBody string
	LogFileDirectory string

	ExecDurationHeaderName string
	CacheUsageHeaderName   string
}


type ConfigReader interface {
	LoadConfig(vars ...interface{}) (*Config, error)
	ParseAssertions() (*assertions.AssertionManager, error)
	ParseHeaders() (http.Header, error)
}

func NewConfig(configType string) ConfigReader {
	if configType == "yml" {

	} else if configType == "cli" {
		return NewConfigCli()
	}
	return nil
}

