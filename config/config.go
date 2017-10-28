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

// UnmarshalYAML unmarshals and assigns names to options and tasks.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {

	type configType Config // Use new type to avoid recursion
	if err := unmarshal((*configType)(c)); err != nil {
		return err
	}

	for name, opt := range c.Options {
		opt.Name = name
	}

	for name, t := range c.Tasks {
		t.Name = name
	}

	return nil
}

// Metadata contains global configuration settings.
type Metadata struct {
	CfgText      []byte
	Directory    string
	PrintHelp    bool
	PrintVersion bool
	Quiet        bool
	Verbose      bool
}
