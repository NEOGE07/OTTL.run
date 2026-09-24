package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads an OpenTelemetry Collector YAML configuration.
func Load(path string) (*CollectorConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg CollectorConfig

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
