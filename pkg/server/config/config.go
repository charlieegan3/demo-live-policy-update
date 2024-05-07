package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Address string         `yaml:"address"`
	Port    int            `yaml:"port"`
	OPAs    map[string]OPA `yaml:"opas"`
}

type OPA struct {
	Endpoint string `yaml:"endpoint"`
	Token    string `yaml:"token"`
	SystemID string `yaml:"system_id"`
}

func ParseConfig(rawConfig []byte) (*Config, error) {
	cfg := &Config{}
	err := yaml.Unmarshal(rawConfig, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}
