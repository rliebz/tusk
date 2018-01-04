package config

import (
	"github.com/rliebz/tusk/config/option"
	"github.com/rliebz/tusk/config/task"
	"github.com/rliebz/tusk/ui"
)

// Config is a struct representing the format for configuration settings.
type Config struct {
	Options map[string]*option.Option
	Tasks   map[string]*task.Task
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
	Verbosity    ui.VerbosityLevel
}
