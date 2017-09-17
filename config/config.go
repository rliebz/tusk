package config

import (
	"github.com/rliebz/tusk/config/option"
	"github.com/rliebz/tusk/config/task"
	yaml "gopkg.in/yaml.v2"
)

// Config is a struct representing the format for configuration settings.
type Config struct {
	Options map[string]*option.Option
	Tasks   map[string]*task.Task
}

// Metadata contains global configuration settings.
type Metadata struct {
	CfgText    []byte
	Directory  string
	RunVersion bool
	Quiet      bool
	Verbose    bool
}

// New is the constructor for Config.
func New() *Config {
	return &Config{
		Options: make(map[string]*option.Option),
		Tasks:   make(map[string]*task.Task),
	}
}

// Parse loads the contents of a config file into a struct.
func Parse(text []byte) (*Config, error) {
	cfg := New()

	if err := yaml.UnmarshalStrict(text, &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
