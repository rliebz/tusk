package task

import (
	"fmt"
	"os"

	"github.com/rliebz/tusk/config/option"
	"github.com/rliebz/tusk/ui"
	yaml "gopkg.in/yaml.v2"
)

// executionState indicates whether a task is "running" or "finally".
type executionState int

const (
	stateRunning executionState = iota
	stateFinally executionState = iota
)

// Task is a single task to be run by CLI.
type Task struct {
	ArgMapSlice    yaml.MapSlice `yaml:"args,omitempty"`
	OptionMapSlice yaml.MapSlice `yaml:"options,omitempty"`

	RunList     RunList `yaml:"run"`
	Finally     RunList `yaml:"finally,omitempty"`
	Usage       string  `yaml:",omitempty"`
	Description string  `yaml:",omitempty"`
	Private     bool

	// Computed members not specified in yaml file
	Name               string                    `yaml:"-"`
	Vars               map[string]string         `yaml:"-"`
	Args               map[string]*option.Arg    `yaml:"-"`
	OrderedArgNames    []string                  `yaml:"-"`
	Options            map[string]*option.Option `yaml:"-"`
	OrderedOptionNames []string                  `yaml:"-"`
}

// UnmarshalYAML unmarshals and assigns names to options.
func (t *Task) UnmarshalYAML(unmarshal func(interface{}) error) error {

	type taskType Task // Use new type to avoid recursion
	if err := unmarshal((*taskType)(t)); err != nil {
		return err
	}

	args, orderedArgs, err := option.GetArgsWithOrder(t.ArgMapSlice)
	if err != nil {
		return err
	}

	t.Args = args
	t.OrderedArgNames = orderedArgs

	options, orderedOptions, err := option.GetOptionsWithOrder(t.OptionMapSlice)
	if err != nil {
		return err
	}

	t.Options = options
	t.OrderedOptionNames = orderedOptions

	for name := range t.Options {
		if _, ok := t.Args[name]; ok {
			return fmt.Errorf(
				"argument and option %q must have unique names for task %q",
				name, t.Name,
			)
		}
	}

	return nil
}

// AllRunItems returns all run items referenced, including `run` and `finally`.
func (t *Task) AllRunItems() RunList {
	return append(t.RunList, t.Finally...)
}

// Dependencies returns a list of options that are required explicitly.
// This does not include interpolations.
func (t *Task) Dependencies() []string {
	options := make([]string, 0, len(t.Options)+len(t.AllRunItems()))

	for _, opt := range t.Options {
		options = append(options, opt.Dependencies()...)
	}
	for _, run := range t.AllRunItems() {
		options = append(options, run.When.Dependencies()...)
	}

	return options
}

// Execute runs the Run scripts in the task.
func (t *Task) Execute(asSubTask bool) (err error) {
	ui.PrintTask(t.Name, asSubTask)

	defer ui.PrintTaskCompleted(t.Name, asSubTask)
	defer t.runFinally(&err, asSubTask)

	for _, r := range t.RunList {
		if rerr := t.run(r, stateRunning); rerr != nil {
			return rerr
		}
	}

	return err
}

func (t *Task) runFinally(err *error, asSubTask bool) {
	if len(t.Finally) == 0 {
		return
	}

	ui.PrintTaskFinally(t.Name, asSubTask)

	for _, r := range t.Finally {
		if rerr := t.run(r, stateFinally); rerr != nil {
			// Do not overwrite existing errors
			if *err == nil {
				*err = rerr
			}
			return
		}
	}
}

// run executes a Run struct.
func (t *Task) run(r *Run, s executionState) error {
	if ok, err := r.shouldRun(t.Vars); !ok || err != nil {
		return err
	}

	runFuncs := []func() error{
		func() error { return t.runCommands(r, s) },
		func() error { return t.runSubTasks(r) },
		func() error { return t.runEnvironment(r) },
	}

	for i := range runFuncs {
		if err := runFuncs[i](); err != nil {
			return err
		}
	}

	return nil
}

func (t *Task) runCommands(r *Run, s executionState) error {
	for _, command := range r.Command {
		switch s {
		case stateFinally:
			ui.PrintCommandWithParenthetical(command, t.Name, "finally")
		default:
			ui.PrintCommand(command, t.Name)
		}

		if err := execCommand(command); err != nil {
			ui.PrintCommandError(err)
			return err
		}
	}

	return nil
}

func (t *Task) runSubTasks(r *Run) error {
	for _, subTask := range r.Tasks {
		if err := subTask.Execute(true); err != nil {
			return err
		}
	}

	return nil
}

func (t *Task) runEnvironment(r *Run) error {
	ui.PrintEnvironment(r.SetEnvironment)
	for key, value := range r.SetEnvironment {
		if value == nil {
			if err := os.Unsetenv(key); err != nil {
				return err
			}

			continue
		}

		if err := os.Setenv(key, *value); err != nil {
			return err
		}
	}

	return nil
}
