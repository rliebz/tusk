package config

import (
	"gitlab.com/rliebz/tusk/task"
	yaml "gopkg.in/yaml.v2"
)

// Config is a struct representing the format for configuration settings.
type Config struct {
	Args  map[string]*task.Arg
	Tasks map[string]*task.Task
}

// New is the constructor for Config.
func New() *Config {
	return &Config{
		Args:  make(map[string]*task.Arg),
		Tasks: make(map[string]*task.Task),
	}
}

// Parse loads the contents of a config file into a struct.
func Parse(text []byte) (*Config, error) {
	cfg := New()

	if err := yaml.Unmarshal(text, &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
