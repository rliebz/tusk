package config

import (
	"fmt"

	"github.com/rliebz/tusk/config/configyaml/option"
	"github.com/rliebz/tusk/config/configyaml/task"
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

	for _, optName := range ordered {
		for _, r := range required {
			if r != optName {
				continue
			}

			value, err := getOptValue(cfgText, passed, options, optName, taskName)
			if err != nil {
				return nil, nil, err
			}

			options[optName] = value

			cfgText, err = interp.Interpolate(cfgText, optName, value)
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

func getOptValue(
	cfgText []byte,
	passed map[string]string,
	options map[string]string,
	optName string,
	taskName string,
) (string, error) {

	cfg, err := Parse(cfgText)
	if err != nil {
		return "", err
	}

	t, ok := cfg.Tasks[taskName]
	if !ok {
		return "", fmt.Errorf(`could not find task "%s"`, taskName)
	}

	if err = AddSubTasks(cfg, t); err != nil {
		return "", err
	}

	opt, err := getOpt(cfg, optName, taskName)
	if err != nil {
		return "", err
	}

	opt.Vars = options

	valuePassed, ok := passed[optName]
	if ok {
		opt.Passed = valuePassed
	}

	return opt.Value()
}

// getOpt gets an option from a Config by name. Task-specific options, sub-
// task options, and global options are checked, in that order.
func getOpt(cfg *Config, optName string, taskName string) (*option.Option, error) {

	if t, ok := cfg.Tasks[taskName]; ok {
		if value, found := checkTaskForOpt(t, optName); found {
			return value, nil
		}
	}

	if value, ok := cfg.Options[optName]; ok {
		return value, nil
	}

	return nil, fmt.Errorf(`option "%s" required but not defined`, optName)
}

// checkTaskForOpt checks a task and its sub-tasks for an option.
func checkTaskForOpt(t *task.Task, optName string) (*option.Option, bool) {

	if value, ok := t.Options[optName]; ok {
		return value, true
	}

	for _, subTask := range t.SubTasks {
		if opt, found := checkTaskForOpt(subTask, optName); found {
			return opt, true
		}
	}

	return nil, false
}
