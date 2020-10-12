package config

import (
	"errors"
	"github.com/go-yaml/yaml"
	"github.com/mostafatalebi/loadtest/pkg/assertions"
	"github.com/mostafatalebi/loadtest/pkg/logger"
	variable "github.com/mostafatalebi/loadtest/pkg/variables"
	"io/ioutil"
	"net/http"
)

type YamlConfigHolder struct {
	Logs        *YamlConfigSectionLogs              `yaml:"logs"`
	Main        *YamlConfigSectionMain              `yaml:"main"`
	DataSources map[string]*YamlConfigSectionTarget `yaml:"data-sources"`
	Targets     map[string]*YamlConfigSectionTarget `yaml:"targets"`
}

type YamlConfigSectionMain struct {
	Concurrency      int64  `yaml:"concurrency"`
	NumberOfRequests int64  `yaml:"request-count"`
	TargetingPolicy  string `yaml:"targeting-policy"`
}

type YamlConfigRefresh struct {
	RefreshType string `yaml:"type"`
	Value       int    `yaml:"value"`
}

type YamlConfigSectionLogs struct {
	Enabled bool   `yaml:"enabled"`
	Dir     string `yaml:"dir"`
}

type YamlConfigTargets map[string]*YamlConfigSectionTarget

type YamlConfigSectionTarget struct {
	Assertions             map[string]string    `yaml:"assertions"`
	Headers                map[string]string    `yaml:"headers"`
	Method                 string               `yaml:"method"`
	Url                    string               `yaml:"url"`
	MaxTimeout             int                  `yaml:"max-timeout"`
	EnabledLogs            bool                 `yaml:"enable-logs"`
	FormBody               string               `yaml:"form-body"`
	LogFileDirectory       string               `yaml:"log-dir"`
	ExecDurationHeaderName string               `yaml:"exec-duration-header-name"`
	CacheUsageHeaderName   string               `yaml:"cache-usage-header-name"`
	Variables              variable.VariableMap `yaml:"variables"`
	TargetingPolicy        string               `yaml:"-"`
	Refresh                *YamlConfigRefresh   `yaml:"refresh"`
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
			if unconvertedConfig != nil {
				cc, err := c.mapYmlToConfig(targetName, unconvertedConfig, c.yamlConfig.Logs)
				if err != nil {
					logger.InfoOut("config failed", err.Error())
					continue
				}
				configs = append(configs, cc)
			}
		}
	}
	if c.yamlConfig != nil && c.yamlConfig.DataSources != nil && len(c.yamlConfig.DataSources) > 0 {
		for targetName, unconvertedConfig := range c.yamlConfig.DataSources {
			if unconvertedConfig != nil {
				cc, err := c.mapYmlToConfig(targetName, unconvertedConfig, c.yamlConfig.Logs)
				if err != nil {
					logger.InfoOut("config failed", err.Error())
					continue
				}
				configs = append(configs, cc)
			}
		}
	}
	return configs, nil
}

func (c *ConfigYaml) mapYmlToConfig(targetName string, ymlConfig *YamlConfigSectionTarget, logsConfig *YamlConfigSectionLogs) (*Config, error) {
	cc := &Config{}
	var err error
	cc.VariablesMap = ymlConfig.Variables
	cc.Assertions, err = c.ParseAssertions(ymlConfig.Assertions)
	if err != nil {
		return nil, err
	}
	cc.Headers, err = c.ParseHeaders(ymlConfig.Headers)
	if err != nil {
		return nil, err
	}
	cc.NumberOfRequests = c.yamlConfig.Main.NumberOfRequests
	cc.Concurrency = c.yamlConfig.Main.Concurrency
	cc.FormBody = ymlConfig.FormBody
	cc.Method = ymlConfig.Method
	cc.Url = ymlConfig.Url
	cc.TargetName = targetName
	cc.MaxTimeout = ymlConfig.MaxTimeout
	if logsConfig != nil {
		cc.EnabledLogs = logsConfig.Enabled
		cc.LogFileDirectory = logsConfig.Dir
	}
	cc.ExecDurationHeaderName = ymlConfig.ExecDurationHeaderName
	cc.CacheUsageHeaderName = ymlConfig.CacheUsageHeaderName
	cc.TargetingPolicy = c.yamlConfig.Main.TargetingPolicy
	return cc, nil
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
	if valuesMap != nil && len(valuesMap) > 0 {
		var headersMap = http.Header{}
		for k, v := range valuesMap {
			headersMap.Set(k, v)
		}
		return headersMap, nil
	}
	return nil, nil
}
