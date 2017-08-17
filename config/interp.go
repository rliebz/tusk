package config

import (
	"fmt"

	"gitlab.com/rliebz/tusk/interp"
	"gitlab.com/rliebz/tusk/task"
	yaml "gopkg.in/yaml.v2"
)

// Interpolate evaluates the variables given and returns interpolated text.
func Interpolate(cfgText []byte, flags map[string]string) ([]byte, error) {

	ordered, err := getOrderedArgs(cfgText)
	if err != nil {
		return nil, err
	}

	for _, name := range ordered {
		cfgText, err = interpolateFlag(cfgText, flags, name)
		if err != nil {
			return nil, err
		}
	}

	return cfgText, nil
}

// getOrderedArgs returns a list of args in the order they appear.
func getOrderedArgs(cfgText []byte) ([]string, error) {

	ordered := new(struct {
		Args  yaml.MapSlice
		Tasks yaml.MapSlice
	})

	if err := yaml.Unmarshal(cfgText, ordered); err != nil {
		return nil, err
	}

	allArgs := ordered.Args

	for _, mapslice := range ordered.Tasks {
		for _, ms := range mapslice.Value.(yaml.MapSlice) {
			name := ms.Key.(string)
			if name != "args" {
				continue
			}

			allArgs = append(allArgs, ms.Value.(yaml.MapSlice)...)
		}
	}

	var output []string
	for _, mapslice := range allArgs {
		name, ok := mapslice.Key.(string)
		if !ok {
			return nil, fmt.Errorf("failed to assert name as string: %v", mapslice.Key)
		}

		output = append(output, name)
	}

	return output, nil
}

// interpolateFlag runs interpolation over config text for a given flag name.
func interpolateFlag(cfgText []byte, flags map[string]string, name string) ([]byte, error) {

	cfg := New()

	if err := yaml.Unmarshal(cfgText, &cfg); err != nil {
		return nil, err
	}

	arg, err := getArg(cfg, name)
	if err != nil {
		return nil, err
	}

	passed, ok := flags[name]
	if ok {
		arg.Passed = passed
	}

	value, err := arg.Value()
	if err != nil {
		return nil, err
	}

	return interp.Interpolate(cfgText, name, value)
}

func getArg(cfg *Config, name string) (*task.Arg, error) {

	if value, ok := cfg.Args[name]; ok {
		return value, nil
	}

	// TODO: Can we limit which tasks we check at this point?
	for _, t := range cfg.Tasks {
		if value, ok := t.Args[name]; ok {
			return value, nil
		}
	}

	return nil, fmt.Errorf("could not find arg %s", name)
}
