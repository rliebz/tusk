package config

import (
	"fmt"

	"github.com/rliebz/tusk/config/option"
	"github.com/rliebz/tusk/config/task"
	"github.com/rliebz/tusk/interp"
	yaml "gopkg.in/yaml.v2"
)

// Parse loads the contents of a config file into a struct.
func Parse(text []byte) (*Config, error) {
	cfg := new(Config)

	if err := yaml.UnmarshalStrict(text, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ParseComplete parses the file completely with interpolation.
func ParseComplete(cfgText []byte, passed map[string]string, taskName string) (*Config, error) {

	cfg, err := Parse(cfgText)
	if err != nil {
		return nil, err
	}

	t, isTaskSet := cfg.Tasks[taskName]
	if !isTaskSet {
		return cfg, nil
	}

	if err := passTaskValues(t, cfg, cfgText, passed); err != nil {
		return nil, err
	}

	return cfg, nil
}

func passTaskValues(t *task.Task, cfg *Config, cfgText []byte, passed map[string]string) error {
	vars, err := interpolateGlobalOptions(t, cfg, cfgText, passed)
	if err != nil {
		return err
	}

	if err := interpolateTask(t, cfgText, passed, vars); err != nil {
		return err
	}

	return addSubTasks(t, cfg, cfgText)
}

func interpolateGlobalOptions(
	t *task.Task, cfg *Config, cfgText []byte, passed map[string]string,
) (map[string]string, error) {

	globalOptions, err := getRequiredGlobalOptions(t, cfgText)
	if err != nil {
		return nil, err
	}

	vars := make(map[string]string, len(globalOptions))
	for _, name := range globalOptions {
		o := cfg.Options[name]
		if err := interpolateOption(o, passed, vars); err != nil {
			return nil, err
		}
	}

	return vars, nil
}

func getRequiredGlobalOptions(t *task.Task, cfgText []byte) ([]string, error) {
	ordered, err := getOrderedGlobalOptions(cfgText)
	if err != nil {
		return nil, err
	}

	cfg, err := Parse(cfgText)
	if err != nil {
		return nil, err
	}

	required, err := FindAllOptions(t, cfg)
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

func getOrderedGlobalOptions(cfgText []byte) ([]string, error) {
	cfgMapSlice := new(struct {
		Options yaml.MapSlice
	})

	if err := yaml.Unmarshal(cfgText, cfgMapSlice); err != nil {
		return nil, err
	}

	var ordered []string
	for _, optionMapSlice := range cfgMapSlice.Options {
		ordered = append(ordered, optionMapSlice.Key.(string))
	}

	return ordered, nil
}

func interpolateOption(o *option.Option, passed, vars map[string]string) error {
	if err := interp.Marshallable(o, vars); err != nil {
		return err
	}

	if valuePassed, ok := passed[o.Name]; ok {
		o.Passed = valuePassed
	}

	o.Vars = vars
	value, err := o.Evaluate()
	if err != nil {
		return err
	}

	vars[o.Name] = value

	return nil
}

func interpolateTask(t *task.Task, cfgText []byte, passed, vars map[string]string) error {
	taskOptions, err := getOrderedTaskOptions(cfgText, t.Name)
	if err != nil {
		return err
	}

	taskVars := make(map[string]string, len(vars)+len(taskOptions))
	for k, v := range vars {
		taskVars[k] = v
	}

	for _, name := range taskOptions {
		o := t.Options[name]

		if err := interpolateOption(o, passed, taskVars); err != nil {
			return err
		}
	}

	if err := interp.Marshallable(&t.RunList, taskVars); err != nil {
		return err
	}

	t.Vars = taskVars

	return nil
}

func getOrderedTaskOptions(cfgText []byte, taskName string) ([]string, error) {
	cfgMapSlice := new(struct {
		Tasks yaml.MapSlice
	})

	if err := yaml.Unmarshal(cfgText, cfgMapSlice); err != nil {
		return nil, err
	}

	var ordered []string
	for _, taskMapSlice := range cfgMapSlice.Tasks {
		if name := taskMapSlice.Key.(string); name != taskName {
			continue
		}

		for _, mapSlice := range taskMapSlice.Value.(yaml.MapSlice) {
			if name := mapSlice.Key.(string); name != "options" {
				continue
			}

			for _, optionMapSlice := range mapSlice.Value.(yaml.MapSlice) {
				ordered = append(ordered, optionMapSlice.Key.(string))
			}
			break
		}
	}

	return ordered, nil
}

func addSubTasks(t *task.Task, cfg *Config, cfgText []byte) error {

	for _, run := range t.RunList {
		for _, subTaskDesc := range run.SubTaskList {
			st, ok := cfg.Tasks[subTaskDesc.Name]
			if !ok {
				return fmt.Errorf(
					`sub-task "%s" does not exist`,
					subTaskDesc.Name,
				)
			}

			passed := subTaskDesc.Options
			subTask := *st

			for optName := range passed {
				if _, isValidOption := subTask.Options[optName]; !isValidOption {
					return fmt.Errorf(
						`option "%s" cannot be passed to task "%s"`,
						optName, subTask.Name,
					)
				}
			}

			if err := passTaskValues(&subTask, cfg, cfgText, passed); err != nil {
				return err
			}

			run.Tasks = append(run.Tasks, subTask)

		}
	}

	return nil
}
