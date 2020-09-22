package config

import (
	_ "github.com/go-yaml/yaml"
)

type ConfigYaml struct {
	rawArgs []string
}

func NewConfigYaml() *ConfigYaml {
	return &ConfigYaml{}
}

func (c *ConfigYaml) LoadConfig(vars ...interface{}) (*Config, error) {
	return nil, nil
}