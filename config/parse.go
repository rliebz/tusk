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
func ParseComplete(
	cfgText []byte,
	taskName string,
	args []string,
	flags map[string]string,
) (*Config, error) {

	cfg, err := Parse(cfgText)
	if err != nil {
		return nil, err
	}

	t, isTaskSet := cfg.Tasks[taskName]
	if !isTaskSet {
		return cfg, nil
	}

	passed, err := combineArgsAndFlags(t, args, flags)
	if err != nil {
		return nil, err
	}

	if err := passTaskValues(t, cfg, passed); err != nil {
		return nil, err
	}

	return cfg, nil
}

func combineArgsAndFlags(
	t *task.Task, args []string, flags map[string]string,
) (map[string]string, error) {
	if len(t.Args) != len(args) {
		return nil, fmt.Errorf(
			"task %q requires exactly %d args, got %d",
			t.Name, len(t.Args), len(args),
		)
	}

	passed := make(map[string]string, len(args)+len(flags))
	i := 0
	for name := range t.Args {
		passed[name] = args[i]
		i++
	}
	for name, value := range flags {
		passed[name] = value
	}

	return passed, nil
}

func passTaskValues(t *task.Task, cfg *Config, passed map[string]string) error {
	vars, err := interpolateGlobalOptions(t, cfg, passed)
	if err != nil {
		return err
	}

	if err := interpolateTask(t, passed, vars); err != nil {
		return err
	}

	return addSubTasks(t, cfg)
}

func interpolateGlobalOptions(
	t *task.Task, cfg *Config, passed map[string]string,
) (map[string]string, error) {

	globalOptions, err := getRequiredGlobalOptions(t, cfg)
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

func getRequiredGlobalOptions(t *task.Task, cfg *Config) ([]string, error) {
	required, err := FindAllOptions(t, cfg)
	if err != nil {
		return nil, err
	}

	var output []string
	for _, o := range cfg.OrderedOptionNames {
		for _, r := range required {
			if r.Name != o {
				continue
			}

			output = append(output, o)
		}
	}

	return output, nil
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

func interpolateTask(t *task.Task, passed, vars map[string]string) error {

	taskVars := make(map[string]string, len(vars)+len(t.Args)+len(t.Options))
	for k, v := range vars {
		taskVars[k] = v
	}

	for _, name := range t.OrderedArgNames {
		a := t.Args[name]

		if err := interpolateArg(a, passed, taskVars); err != nil {
			return err
		}
	}

	for _, name := range t.OrderedOptionNames {
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

func addSubTasks(t *task.Task, cfg *Config) error {

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

			values, err := getArgValues(subTask, subTaskDesc.Args)
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

			if err := passTaskValues(subTask, cfg, values); err != nil {
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
	subTask *task.Task, argsPassed []string,
) (map[string]string, error) {

	if len(argsPassed) != len(subTask.Args) {
		return nil, fmt.Errorf(
			"subtask %q requires %d args but got %d",
			subTask.Name, len(subTask.Args), len(argsPassed),
		)
	}

	values := make(map[string]string)
	for i, argName := range subTask.OrderedArgNames {
		values[argName] = argsPassed[i]
	}

	return values, nil

}
