package runner

import (
	"fmt"

	"github.com/rliebz/tusk/marshal"
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
	t *Task, args []string, flags map[string]string,
) (map[string]string, error) {
	if len(t.Args) != len(args) {
		return nil, fmt.Errorf(
			"task %q requires exactly %d args, got %d",
			t.Name, len(t.Args), len(args),
		)
	}

	passed := make(map[string]string, len(args)+len(flags))
	for i, arg := range t.Args {
		passed[arg.Name] = args[i]
	}
	for name, value := range flags {
		passed[name] = value
	}

	return passed, nil
}

func passTaskValues(t *Task, cfg *Config, passed map[string]string) error {
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
	t *Task, cfg *Config, passed map[string]string,
) (map[string]string, error) {
	globalOptions, err := getRequiredGlobalOptions(t, cfg)
	if err != nil {
		return nil, err
	}

	vars := make(map[string]string, len(globalOptions))
	for _, o := range globalOptions {
		if err := interpolateOption(o, passed, vars); err != nil {
			return nil, err
		}
	}

	return vars, nil
}

func getRequiredGlobalOptions(t *Task, cfg *Config) (Options, error) {
	required, err := FindAllOptions(t, cfg)
	if err != nil {
		return nil, err
	}

	var output Options
	for _, o := range cfg.Options {
		for _, r := range required {
			if r.Name != o.Name {
				continue
			}

			output = append(output, o)
		}
	}

	return output, nil
}

func interpolateArg(a *Arg, passed, vars map[string]string) error {
	if err := marshal.Interpolate(a, vars); err != nil {
		return err
	}

	valuePassed, ok := passed[a.Name]
	if !ok {
		return fmt.Errorf("no value passed for arg %q", a.Name)
	}

	a.Passed = valuePassed

	value, err := a.Evaluate()
	if err != nil {
		return err
	}

	vars[a.Name] = value

	return nil
}

func interpolateOption(o *Option, passed, vars map[string]string) error {
	if err := marshal.Interpolate(o, vars); err != nil {
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

func interpolateTask(t *Task, passed, vars map[string]string) error {
	taskVars := make(map[string]string, len(vars)+len(t.Args)+len(t.Options))
	for k, v := range vars {
		taskVars[k] = v
	}

	for _, a := range t.Args {
		if err := interpolateArg(a, passed, taskVars); err != nil {
			return err
		}
	}

	for _, o := range t.Options {
		if err := interpolateOption(o, passed, taskVars); err != nil {
			return err
		}
	}

	if err := marshal.Interpolate(&t.RunList, taskVars); err != nil {
		return err
	}

	if err := marshal.Interpolate(&t.Finally, taskVars); err != nil {
		return err
	}

	t.Vars = taskVars

	return nil
}

func addSubTasks(t *Task, cfg *Config) error {
	for _, run := range t.AllRunItems() {
		for _, desc := range run.SubTaskList {
			sub, err := newTaskFromSub(desc, cfg)
			if err != nil {
				return err
			}

			run.Tasks = append(run.Tasks, *sub)
		}
	}

	return nil
}

func newTaskFromSub(desc *SubTask, cfg *Config) (*Task, error) {
	st, ok := cfg.Tasks[desc.Name]
	if !ok {
		return nil, fmt.Errorf("sub-task %q does not exist", desc.Name)
	}

	subTask := copyTask(st)

	values, err := getArgValues(subTask, desc.Args)
	if err != nil {
		return nil, err
	}

	for optName, opt := range desc.Options {
		if _, isValidOption := subTask.Options.Lookup(optName); !isValidOption {
			return nil, fmt.Errorf(
				"option %q cannot be passed to task %q",
				optName, subTask.Name,
			)
		}
		values[optName] = opt
	}

	if err := passTaskValues(subTask, cfg, values); err != nil {
		return nil, err
	}

	return subTask, nil
}

// copyTask returns a copy of a task, replacing references with new values.
func copyTask(t *Task) *Task {
	newTask := *t

	argsCopy := make(Args, 0, len(newTask.Args))
	for _, ptr := range newTask.Args {
		arg := *ptr
		argsCopy = append(argsCopy, &arg)
	}
	newTask.Args = argsCopy

	optionsCopy := make(Options, 0, len(newTask.Options))
	for _, ptr := range newTask.Options {
		opt := *ptr
		optionsCopy = append(optionsCopy, &opt)
	}
	newTask.Options = optionsCopy

	return &newTask
}

func getArgValues(subTask *Task, argsPassed []string) (map[string]string, error) {
	if len(argsPassed) != len(subTask.Args) {
		return nil, fmt.Errorf(
			"subtask %q requires %d args but got %d",
			subTask.Name, len(subTask.Args), len(argsPassed),
		)
	}

	values := make(map[string]string)
	for i, arg := range subTask.Args {
		values[arg.Name] = argsPassed[i]
	}

	return values, nil
}
