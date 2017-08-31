package config

import (
	"fmt"

	"github.com/rliebz/tusk/config/task"
	"github.com/rliebz/tusk/interp"
	yaml "gopkg.in/yaml.v2"
)

// Interpolate evaluates the variables given and returns interpolated text.
//
// cfgText should be a valid, uninterpolated yaml configuration. While
// there is currently no distinct validation phase, it is likely that this
// function would return an error for invalid interpolation syntax.
//
// passed is a map of variable names to values, which are the values of the
// flags that were passed directly by CLI. These will be used in determining
// their own values to interpolate, and also may have an impact on other
// dependent variables that are not overridden by command-line options.
//
// taskName is the name of the task being run. This is used to determine the
// list of options which require interpolation.
func Interpolate(cfgText []byte, passed map[string]string, taskName string) ([]byte, map[string]string, error) {

	options := make(map[string]string)

	ordered, err := getOrderedOpts(cfgText)
	if err != nil {
		return nil, nil, err
	}

	required, err := getRequiredOpts(cfgText, taskName)
	if err != nil {
		return nil, nil, err
	}

	for _, name := range ordered {
		for _, opt := range required {
			if opt != name {
				continue
			}

			value, err := getFlagValue(cfgText, passed, options, name)
			if err != nil {
				return nil, nil, err
			}

			options[name] = value

			cfgText, err = interp.Interpolate(cfgText, name, value)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	// TODO: Can this return only options?
	return cfgText, options, nil
}

func getRequiredOpts(cfgText []byte, taskName string) ([]string, error) {
	if taskName == "" {
		return []string{}, nil
	}

	cfg, err := Parse(cfgText)
	if err != nil {
		return nil, err
	}

	t, ok := cfg.Tasks[taskName]
	if !ok {
		return nil, fmt.Errorf("could not find task `%s`", taskName)
	}

	if err = AddSubTasks(cfg, t); err != nil {
		return nil, err
	}

	required, err := cfg.FindAllOptions(t)
	if err != nil {
		return nil, err
	}

	var output []string
	for _, opt := range required {
		output = append(output, opt.Name)
	}

	return output, nil
}

// getOrderedOpts returns a list of options in the order they appear.
func getOrderedOpts(cfgText []byte) ([]string, error) {

	ordered := new(struct {
		Options yaml.MapSlice
		Tasks   yaml.MapSlice
	})

	if err := yaml.Unmarshal(cfgText, ordered); err != nil {
		return nil, err
	}

	allOpts := ordered.Options

	for _, mapslice := range ordered.Tasks {
		for _, ms := range mapslice.Value.(yaml.MapSlice) {
			name := ms.Key.(string)
			if name != "options" {
				continue
			}

			allOpts = append(allOpts, ms.Value.(yaml.MapSlice)...)
		}
	}

	var output []string
	for _, mapslice := range allOpts {
		name, ok := mapslice.Key.(string)
		if !ok {
			return nil, fmt.Errorf("failed to assert name as string: %v", mapslice.Key)
		}

		output = append(output, name)
	}

	return output, nil
}

func getFlagValue(cfgText []byte, passed map[string]string, options map[string]string, name string) (string, error) {

	cfg := New()

	if err := yaml.Unmarshal(cfgText, &cfg); err != nil {
		return "", err
	}

	opt, err := getOpt(cfg, name)
	if err != nil {
		return "", err
	}

	opt.Vars = options

	valuePassed, ok := passed[name]
	if ok {
		opt.Passed = valuePassed
	}

	return opt.Value()
}

// getOpt gets an option from a Config by name. Both global options and
// task-specific options are checked.
func getOpt(cfg *Config, name string) (*task.Option, error) {

	if value, ok := cfg.Options[name]; ok {
		return value, nil
	}

	// TODO: Can we limit which tasks we check at this point?
	for _, t := range cfg.Tasks {
		if value, ok := t.Options[name]; ok {
			return value, nil
		}
	}

	return nil, fmt.Errorf("option \"%s\" required but not defined", name)
}
