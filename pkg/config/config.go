package config

const (
	FieldExecDurationHeaderName = "exec-duration-header-name"
	FieldCacheUsageHeaderName   = "cache-usage-header-name"
	FieldConcurrency            = "concurrency"
	FieldNumberOfRequests       = "request-count"
	FieldMethod                 = "method"
	FieldUrl                    = "url"
	FieldMaxTimeout             = "max-timeout"
	FieldEnableLogs             = "enable-logs"
	FieldAssertBodyString       = "assert-body-string"
)


type Config struct {
	Concurrency int64
	NumberOfRequests int64
	Method string
	Url string
	MaxTimeout int
	EnabledLogs bool
	Assertions map[string]interface{}

	ExecDurationHeaderName string
	CacheUsageHeaderName   string
}


type ConfigReader interface {
	LoadConfig(vars ...interface{}) (*Config, error)
	ParseAssertion() (map[string]interface{}, error)
	ParseHeaders()
}

func NewConfig(configType string) ConfigReader {
	if configType == "yml" {

	} else if configType == "cli" {
		return NewConfigCli()
	}
	return nil
}

