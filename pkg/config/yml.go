package config

import (
	"errors"
	"github.com/go-yaml/yaml"
	"github.com/mostafatalebi/loadtest/pkg/assertions"
	"io/ioutil"
	"net/http"
)

type YamlConfigMain map[string]*YamlConfigMap

type YamlConfigMap struct {
	Assertions map[string]string `yaml:"assertions"`
	Headers map[string]string `yaml:"headers"`
	Concurrency int64 `yaml:"concurrency"`
	NumberOfRequests int64 `yaml:"request-count"`
	Method string `yaml:"method"`
	Url string `yaml:"url"`
	MaxTimeout int `yaml:"max-timeout"`
	EnabledLogs bool `yaml:"enable-logs"`
	FormBody string `yaml:"form-body"`
	LogFileDirectory string `yaml:"log-dir"`
	ExecDurationHeaderName string `yaml:"exec-duration-header-name"`
	CacheUsageHeaderName   string `yaml:"cache-usage-header-name"`
}

type ConfigYaml struct {
	rawBytes []byte
	yamlConfig YamlConfigMain
}

func NewConfigYaml() *ConfigYaml {
	return &ConfigYaml{}
}

func (c *ConfigYaml) LoadConfigs(vars ...interface{}) ([]*Config, error) {
	if len(vars) == 0 {
		return nil, errors.New("yaml config file is required")
	}
	var fileName = vars[0].(string)
	b, err := c.readFile(fileName)
	if err != nil {
		return nil, err
	}
	cnf := &Config{}
	err = yaml.Unmarshal(b, cnf)
	if err != nil {
		return nil, err
	}
	c.rawBytes = b
	ymlCnf := &YamlConfigMain{}
	err = yaml.Unmarshal(c.rawBytes, ymlCnf)
	c.yamlConfig = *ymlCnf
	if err != nil {
		return nil, err
	}
	var configs = make([]*Config, 0)
	if c.yamlConfig != nil && len(c.yamlConfig) > 0 {
		for _, unconvertedConfig := range c.yamlConfig {
			if unconvertedConfig != nil {
				cc := &Config{}
				cc.Assertions, err = c.ParseAssertions(unconvertedConfig.Assertions)
				if err != nil {
					return nil, err
				}
				cc.Headers, err = c.ParseHeaders(unconvertedConfig.Headers)
				if err != nil {
					return nil, err
				}
				cc.NumberOfRequests = unconvertedConfig.NumberOfRequests
				cc.Concurrency = unconvertedConfig.Concurrency
				cc.EnabledLogs = unconvertedConfig.EnabledLogs
				cc.LogFileDirectory = unconvertedConfig.LogFileDirectory
				cc.FormBody = unconvertedConfig.FormBody
				cc.Method = unconvertedConfig.Method
				cc.Url = unconvertedConfig.Url
				cc.MaxTimeout = unconvertedConfig.MaxTimeout
				cc.ExecDurationHeaderName = unconvertedConfig.ExecDurationHeaderName
				cc.CacheUsageHeaderName = unconvertedConfig.CacheUsageHeaderName
				configs = append(configs, cc)
			}
		}
	}
	return configs, nil
}

func (c *ConfigYaml) readFile(fileName string) ([]byte, error) {
	return ioutil.ReadFile(fileName)
}

func (c *ConfigYaml) ParseAssertions(valuesMap map[string]string) (*assertions.AssertionManager, error) {
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

func (c *ConfigYaml) ParseHeaders(valuesMap map[string]string) (http.Header, error) {
	if  valuesMap != nil && len(valuesMap) > 0 {
		var headersMap = http.Header{}
		for k, v := range valuesMap {
			headersMap.Set(k, v)
		}
		return headersMap, nil
	}
	return nil, nil
}