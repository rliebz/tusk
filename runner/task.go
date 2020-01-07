package runner

import (
	"fmt"
	"os"

	"github.com/rliebz/tusk/ui"
)

// executionState indicates whether a task is "running" or "finally".
type executionState int

const (
	stateRunning executionState = iota
	stateFinally executionState = iota
)

// Task is a single task to be run by CLI.
type Task struct {
	Args    Args    `yaml:"args,omitempty"`
	Options Options `yaml:"options,omitempty"`

	RunList     RunList `yaml:"run"`
	Finally     RunList `yaml:"finally,omitempty"`
	Usage       string  `yaml:",omitempty"`
	Description string  `yaml:",omitempty"`
	Private     bool

	// Computed members not specified in yaml file
	Name string            `yaml:"-"`
	Vars map[string]string `yaml:"-"`
}

// UnmarshalYAML unmarshals and assigns names to options.
func (t *Task) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type taskType Task // Use new type to avoid recursion
	if err := unmarshal((*taskType)(t)); err != nil {
		return err
	}

	for _, o := range t.Options {
		for _, a := range t.Args {
			if o.Name == a.Name {
				return fmt.Errorf(
					"argument and option %q must have unique names for task %q",
					o.Name, t.Name,
				)
			}
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
func (t *Task) Execute(ctx RunContext) (err error) {
	if !t.Private {
		ctx.PushTask(t)
	}

	ui.PrintTask(t.Name)

	defer ui.PrintTaskCompleted(t.Name)
	defer t.runFinally(ctx, &err)

	for _, r := range t.RunList {
		if rerr := t.run(ctx, r, stateRunning); rerr != nil {
			return rerr
		}
	}

	return err
}

func (t *Task) runFinally(ctx RunContext, err *error) {
	if len(t.Finally) == 0 {
		return
	}

	ui.PrintTaskFinally(t.Name)

	for _, r := range t.Finally {
		if rerr := t.run(ctx, r, stateFinally); rerr != nil {
			// Do not overwrite existing errors
			if *err == nil {
				*err = rerr
			}
			return
		}
	}
}

// run executes a Run struct.
func (t *Task) run(ctx RunContext, r *Run, s executionState) error {
	if ok, err := r.shouldRun(t.Vars); !ok || err != nil {
		return err
	}

	runFuncs := []func() error{
		func() error { return t.runCommands(ctx, r, s) },
		func() error { return t.runSubTasks(ctx, r) },
		func() error { return t.runEnvironment(r) },
	}

	for i := range runFuncs {
		if err := runFuncs[i](); err != nil {
			return err
		}
	}

	return nil
}

func (t *Task) runCommands(ctx RunContext, r *Run, s executionState) error {
	for _, command := range r.Command {
		switch s {
		case stateFinally:
			ui.PrintCommandWithParenthetical(command.Print, "finally", ctx.Tasks()...)
		default:
			ui.PrintCommand(command.Print, ctx.Tasks()...)
		}

		if err := command.exec(); err != nil {
			ui.PrintCommandError(err)
			return err
		}
	}

	return nil
}

func (t *Task) runSubTasks(ctx RunContext, r *Run) error {
	for i := range r.Tasks {
		if err := r.Tasks[i].Execute(ctx); err != nil {
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
