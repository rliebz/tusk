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

	ordered := make([]string, 0, len(cfgMapSlice.Options))
	for _, optionMapSlice := range cfgMapSlice.Options {
		ordered = append(ordered, optionMapSlice.Key.(string))
	}

	return ordered, nil
}

func interpolateArg(a *option.Arg, passed, vars map[string]string) error {
	if err := interp.Marshallable(a, vars); err != nil {
		return err
	}

	if valuePassed, ok := passed[a.Name]; ok {
		a.Passed = valuePassed
	} else {
		return fmt.Errorf("no value passed for arg %q", a.Name)
	}

	value, err := a.Evaluate()
	if err != nil {
		return err
	}

	vars[a.Name] = value

	return nil
}

func interpolateOption(o *option.Option, passed, vars map[string]string) error {
	if err := interp.Marshallable(o, vars); err != nil {
		return err
	}

	if valuePassed, ok := passed[o.Name]; ok {
		o.Passed = valuePassed
	}

	value, err := o.Evaluate(vars)
	if err != nil {
		return err
	}

	vars[o.Name] = value

	return nil
}

func interpolateTask(t *task.Task, cfgText []byte, passed, vars map[string]string) error {

	taskArgs, err := getOrderedTaskOptions(cfgText, t.Name, "args")
	if err != nil {
		return err
	}

	taskOptions, err := getOrderedTaskOptions(cfgText, t.Name, "options")
	if err != nil {
		return err
	}

	taskVars := make(map[string]string, len(vars)+len(taskArgs)+len(taskOptions))
	for k, v := range vars {
		taskVars[k] = v
	}

	for _, name := range taskArgs {
		a := t.Args[name]

		if err := interpolateArg(a, passed, taskVars); err != nil {
			return err
		}
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

func getOrderedTaskOptions(cfgText []byte, taskName, key string) ([]string, error) {
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
			if name := mapSlice.Key.(string); name != key {
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
					"sub-task %q does not exist",
					subTaskDesc.Name,
				)
			}

			subTask := copyTask(st)

			values, err := getArgValues(cfgText, subTask, subTaskDesc.Args)
			if err != nil {
				return err
			}

			for optName, opt := range subTaskDesc.Options {
				if _, isValidOption := subTask.Options[optName]; !isValidOption {
					return fmt.Errorf(
						"option %q cannot be passed to task %q",
						optName, subTask.Name,
					)
				}
				values[optName] = opt
			}

			if err := passTaskValues(subTask, cfg, cfgText, values); err != nil {
				return err
			}

			run.Tasks = append(run.Tasks, *subTask)
		}
	}

	return nil
}

// copyTask returns a copy of a task, replacing references with new values.
func copyTask(t *task.Task) *task.Task {
	newTask := *t

	argsCopy := make(map[string]*option.Arg, len(newTask.Args))
	for name, ptr := range newTask.Args {
		arg := *ptr
		argsCopy[name] = &arg
	}
	newTask.Args = argsCopy

	optionsCopy := make(map[string]*option.Option, len(newTask.Options))
	for name, ptr := range newTask.Options {
		opt := *ptr
		optionsCopy[name] = &opt
	}
	newTask.Options = optionsCopy

	return &newTask
}

func getArgValues(
	cfgText []byte, subTask *task.Task, argsPassed []string,
) (map[string]string, error) {

	if len(argsPassed) != len(subTask.Args) {
		return nil, fmt.Errorf(
			"subtask %q requires %d args but got %d",
			subTask.Name, len(subTask.Args), len(argsPassed),
		)
	}

	values := make(map[string]string)
	subTaskArgs, err := getOrderedTaskOptions(cfgText, subTask.Name, "args")
	if err != nil {
		return nil, err
	}
	for i, argName := range subTaskArgs {
		values[argName] = argsPassed[i]
	}

	return values, nil

}
