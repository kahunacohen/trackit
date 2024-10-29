package config

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

type Account struct {
	Headers map[string]string
}
type Config struct {
	Accounts   map[string]Account
	Categories []string
	Data       string
}

func ParseConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file at %s: %w", path, err)
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}
	return &config, nil
}
