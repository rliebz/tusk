package runner

import (
	"fmt"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"github.com/rliebz/tusk/marshal"
)

// Parse loads the contents of a config file into a struct.
func Parse(text []byte) (*Config, error) {
	var cfg Config
	if err := yaml.UnmarshalStrict(text, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// ParseComplete parses the file completely with env file parsing and
// interpolation.
func ParseComplete(
	meta *Metadata,
	taskName string,
	args []string,
	flags map[string]string,
) (*Config, error) {
	cfg, err := Parse(meta.CfgText)
	if err != nil {
		return nil, err
	}

	err = loadEnvFiles(filepath.Dir(meta.CfgPath), cfg.EnvFile)
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

	ctx := Context{
		CfgPath:     meta.CfgPath,
		Interpreter: meta.Interpreter,
	}

	if err := passTaskValues(ctx, t, cfg, passed); err != nil {
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

func passTaskValues(
	ctx Context,
	t *Task,
	cfg *Config,
	passed map[string]string,
) error {
	vars, err := interpolateGlobalOptions(ctx, t, cfg, passed)
	if err != nil {
		return err
	}

	if err := interpolateTask(ctx, t, passed, vars); err != nil {
		return err
	}

	return addSubTasks(ctx, t, cfg)
}

func interpolateGlobalOptions(
	ctx Context,
	t *Task,
	cfg *Config,
	passed map[string]string,
) (map[string]string, error) {
	globalOptions, err := getRequiredGlobalOptions(t, cfg)
	if err != nil {
		return nil, err
	}

	vars := make(map[string]string, len(globalOptions))
	for _, o := range globalOptions {
		if err := interpolateOption(ctx, o, passed, vars); err != nil {
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

func interpolateOption(ctx Context, o *Option, passed, vars map[string]string) error {
	if err := marshal.Interpolate(o, vars); err != nil {
		return err
	}

	if valuePassed, ok := passed[o.Name]; ok {
		o.Passed = valuePassed
	}

	value, err := o.Evaluate(ctx, vars)
	if err != nil {
		return err
	}

	if o.isBoolean() && o.Rewrite != "" {
		switch value {
		case "true":
			value = o.Rewrite
		case "false":
			value = ""
		}
	}

	vars[o.Name] = value

	return nil
}

func interpolateTask(ctx Context, t *Task, passed, vars map[string]string) error {
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
		if err := interpolateOption(ctx, o, passed, taskVars); err != nil {
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

func addSubTasks(ctx Context, t *Task, cfg *Config) error {
	for _, run := range t.AllRunItems() {
		for _, desc := range run.SubTaskList {
			sub, err := newTaskFromSub(ctx, desc, cfg)
			if err != nil {
				return err
			}

			run.Tasks = append(run.Tasks, *sub)
		}
	}

	return nil
}

func newTaskFromSub(ctx Context, desc *SubTask, cfg *Config) (*Task, error) {
	st, ok := cfg.Tasks[desc.Name]
	if !ok {
		return nil, fmt.Errorf("sub-task %q is not defined", desc.Name)
	}

	subTask := copyTask(st)

	values, err := getArgValues(subTask, desc.Args)
	if err != nil {
		return nil, err
	}

	for optName, optValue := range desc.Options {
		opt, ok := subTask.Options.Lookup(optName)
		if !ok {
			return nil, fmt.Errorf(
				"option %q cannot be passed to task %q",
				optName, subTask.Name,
			)
		}

		if err := opt.validatePassed(optValue); err != nil {
			return nil, err
		}

		values[optName] = optValue
	}

	if err := passTaskValues(ctx, subTask, cfg, values); err != nil {
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
		argValue := argsPassed[i]
		if err := arg.validatePassed(argValue); err != nil {
			return nil, err
		}

		values[arg.Name] = argValue
	}

	return values, nil
}
