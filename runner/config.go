package runner

// Config is a struct representing the format for configuration settings.
type Config struct {
	Name  string `yaml:"name"`
	Usage string `yaml:"usage"`

	Tasks   map[string]*Task `yaml:"tasks"`
	Options Options          `yaml:"options,omitempty"`
}

// UnmarshalYAML unmarshals and assigns names to options and tasks.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type configType Config // Use new type to avoid recursion
	if err := unmarshal((*configType)(c)); err != nil {
		return err
	}

	for name, t := range c.Tasks {
		t.Name = name
	}

	return nil
}
