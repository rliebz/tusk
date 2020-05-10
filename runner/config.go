package runner

// Config is a struct representing the format for configuration settings.
type Config struct {
	Name  string `yaml:"name"`
	Usage string `yaml:"usage"`
	// The Interpreter field must be read before the config struct can be parsed
	// completely from YAML. To do so, the config text parses it elsewhere in the
	// code base independently from this struct.
	//
	// It is included here only so that strict unmarshaling does not fail.
	Interpreter string `yaml:"interpreter"`

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
