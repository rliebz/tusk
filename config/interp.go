package config

import (
	"fmt"

	"github.com/rliebz/tusk/config/option"
	"github.com/rliebz/tusk/config/task"
	"github.com/rliebz/tusk/interp"
	yaml "gopkg.in/yaml.v2"
)

// ParseComplete parses the file completely with interpolation.
func ParseComplete(cfgText []byte, passed map[string]string, taskName string) (*Config, error) {

	cfg, err := Parse(cfgText)
	if err != nil {
		return nil, err
	}

	t := cfg.Tasks[taskName]

	// TODO: Disallow passing non-options explicitly to subtasks

	values, err := interpolateGlobalOptions(cfg, cfgText, passed, t)
	if err != nil {
		return nil, err
	}

	if t, ok := cfg.Tasks[taskName]; ok {
		if err := interpolateTask(cfgText, values, passed, t); err != nil {
			return nil, err
		}

		if err := addSubTasks(cfg, cfgText, t); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func interpolateGlobalOptions(
	cfg *Config, cfgText []byte, passed map[string]string, t *task.Task,
) (map[string]string, error) {

	globalOptions, err := getRequiredOptions(cfgText, t)
	if err != nil {
		return nil, err
	}

	values := make(map[string]string, len(globalOptions))
	for _, name := range globalOptions {
		o := cfg.Options[name]
		if err := interpolateOption(o, passed, values); err != nil {
			return nil, err
		}
	}

	return values, nil
}

func interpolateTask(cfgText []byte, values, passed map[string]string, t *task.Task) error {
	if t == nil {
		return nil
	}

	taskOptions, err := getTaskOptions(cfgText, t.Name)
	if err != nil {
		return err
	}

	taskValues := make(map[string]string, len(values)+len(taskOptions))
	for k, v := range values {
		taskValues[k] = v
	}

	for _, name := range taskOptions {
		o := t.Options[name]

		o.InvalidateCache()
		if err := interpolateOption(o, passed, taskValues); err != nil {
			return err
		}
	}

	if err := interp.Struct(&t.RunList, taskValues); err != nil {
		return err
	}

	t.Vars = taskValues

	return nil
}

func interpolateOption(o *option.Option, passed, values map[string]string) error {
	if err := interp.Struct(o, values); err != nil {
		return err
	}

	if valuePassed, ok := passed[o.Name]; ok {
		o.Passed = valuePassed
	}

	o.Vars = values
	value, err := o.Evaluate()
	if err != nil {
		return err
	}

	values[o.Name] = value

	return nil
}

func getRequiredOptions(cfgText []byte, t *task.Task) ([]string, error) {
	if t == nil {
		return []string{}, nil
	}

	ordered, err := getOrderedOptions(cfgText)
	if err != nil {
		return nil, err
	}

	cfg, err := Parse(cfgText)
	if err != nil {
		return nil, err
	}

	required, err := cfg.FindAllOptions(t)
	if err != nil {
		return nil, err
	}

	var output []string
	for _, o := range ordered {
		for _, r := range required {
			if r.Name != o {
				continue
			}

			output = append(output, o)
		}
	}

	return output, nil
}

func getOrderedOptions(text []byte) ([]string, error) {
	ordered := new(struct {
		Options yaml.MapSlice
	})

	if err := yaml.Unmarshal(text, ordered); err != nil {
		return nil, err
	}

	var output []string
	for _, mapslice := range ordered.Options {
		name, ok := mapslice.Key.(string)
		if !ok {
			return nil, fmt.Errorf("failed to assert name as string: %v", mapslice.Key)
		}

		output = append(output, name)
	}

	return output, nil
}

func getTaskOptions(cfgText []byte, taskName string) ([]string, error) {
	ordered := new(struct {
		Tasks yaml.MapSlice
	})

	if err := yaml.Unmarshal(cfgText, ordered); err != nil {
		return nil, err
	}

	var output []string
	for _, mapslice := range ordered.Tasks {
		if name := mapslice.Key.(string); name != taskName {
			continue
		}

		for _, mapslice := range mapslice.Value.(yaml.MapSlice) {
			if name := mapslice.Key.(string); name != "options" {
				continue
			}

			for _, mapslice := range mapslice.Value.(yaml.MapSlice) {
				output = append(output, mapslice.Key.(string))
			}
			break
		}
	}

	return output, nil
}
