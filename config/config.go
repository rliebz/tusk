package config

import (
	"github.com/rliebz/tusk/config/option"
	"github.com/rliebz/tusk/config/task"
	yaml "gopkg.in/yaml.v2"
)

// Config is a struct representing the format for configuration settings.
type Config struct {
	Name  string
	Usage string

	Tasks          map[string]*task.Task
	OptionMapSlice yaml.MapSlice `yaml:"options,omitempty"`

	// Computed members not specified in yaml file
	Options            map[string]*option.Option `yaml:"-"`
	OrderedOptionNames []string                  `yaml:"-"`
}

// UnmarshalYAML unmarshals and assigns names to options and tasks.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type configType Config // Use new type to avoid recursion
	if err := unmarshal((*configType)(c)); err != nil {
		return err
	}

	options, ordered, err := option.GetOptionsWithOrder(c.OptionMapSlice)
	if err != nil {
		return err
	}

	c.Options = options
	c.OrderedOptionNames = ordered

	for name, t := range c.Tasks {
		t.Name = name
	}

	return nil
}
