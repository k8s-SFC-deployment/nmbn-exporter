package config

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Targets      []TargetConfig `yaml:"targets"`
	PingInterval int            `yaml:"pingInterval"`
}

type TargetConfig struct {
	IP string `yaml:"ip"`
}

func FromYAML(r io.Reader) (*Config, error) {
	c := &Config{}
	err := yaml.NewDecoder(r).Decode(c)
	if err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}
	return c, nil
}

func ToYAML(w io.Writer, cfg *Config) error {
	err := yaml.NewEncoder(w).Encode(cfg)
	if err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}
	return nil
}
