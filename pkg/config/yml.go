package config

import (
	"errors"
	"github.com/go-yaml/yaml"
	"github.com/mostafatalebi/loadtest/pkg/assertions"
	"io/ioutil"
	"net/http"
)

type YamlConfigHolder struct {
	Main *YamlConfigSectionMain `yaml:"main"`
	Targets map[string]*YamlConfigSectionTarget `yaml:"targets"`
}

type YamlConfigSectionMain struct {
	Concurrency      int64                        `yaml:"concurrency"`
	NumberOfRequests int64                        `yaml:"request-count"`
	DataSource       *YamlConfigSectionDataSource `yaml:"data-source"`
	TargetingPolicy  string                       `yaml:"targeting-policy"`
}

type YamlConfigTargets map[string]*YamlConfigSectionTarget

type YamlConfigSectionTarget struct {
	Assertions map[string]string `yaml:"assertions"`
	Headers map[string]string `yaml:"headers"`
	Method string `yaml:"method"`
	Url string `yaml:"url"`
	MaxTimeout int `yaml:"max-timeout"`
	EnabledLogs bool `yaml:"enable-logs"`
	FormBody string `yaml:"form-body"`
	LogFileDirectory string `yaml:"log-dir"`
	ExecDurationHeaderName string `yaml:"exec-duration-header-name"`
	CacheUsageHeaderName   string `yaml:"cache-usage-header-name"`
	TargetingPolicy string `yaml:"-"`
}

type YamlConfigSectionDataSource struct {
	Concurrency int64 `yaml:"concurrency"`
	NumberOfRequests int64 `yaml:"request-count"`
}

type ConfigYaml struct {
	rawBytes   []byte
	yamlConfig *YamlConfigHolder
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
	ymlCnf := &YamlConfigHolder{}
	err = yaml.Unmarshal(c.rawBytes, ymlCnf)
	c.yamlConfig = ymlCnf
	if err != nil {
		return nil, err
	}
	var configs = make([]*Config, 0)
	if c.yamlConfig != nil && c.yamlConfig.Targets != nil && len(c.yamlConfig.Targets) > 0 {
		for targetName, unconvertedConfig := range c.yamlConfig.Targets {
			if unconvertedConfig.Assertions == nil {}
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
				cc.NumberOfRequests = c.yamlConfig.Main.NumberOfRequests
				cc.Concurrency = c.yamlConfig.Main.Concurrency
				cc.EnabledLogs = unconvertedConfig.EnabledLogs
				cc.LogFileDirectory = unconvertedConfig.LogFileDirectory
				cc.FormBody = unconvertedConfig.FormBody
				cc.Method = unconvertedConfig.Method
				cc.Url = unconvertedConfig.Url
				cc.TargetName = targetName
				cc.MaxTimeout = unconvertedConfig.MaxTimeout
				cc.ExecDurationHeaderName = unconvertedConfig.ExecDurationHeaderName
				cc.CacheUsageHeaderName = unconvertedConfig.CacheUsageHeaderName
				cc.TargetingPolicy = c.yamlConfig.Main.TargetingPolicy
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
	return assertions.NewAssertionManagerWithDefaults(nil), nil
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