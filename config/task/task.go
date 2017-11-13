package task

import (
	"os"

	"github.com/rliebz/tusk/config/option"
	"github.com/rliebz/tusk/config/run"
	"github.com/rliebz/tusk/config/when"
	"github.com/rliebz/tusk/ui"
)

// Task is a single task to be run by CLI.
type Task struct {
	Options     map[string]*option.Option `yaml:",omitempty"`
	Run         run.List
	Usage       string `yaml:",omitempty"`
	Description string `yaml:",omitempty"`
	Private     bool

	// Computed members not specified in yaml file
	Name     string            `yaml:"-"`
	SubTasks []*Task           `yaml:"-"`
	Vars     map[string]string `yaml:"-"`
}

// UnmarshalYAML unmarshals and assigns names to options.
func (t *Task) UnmarshalYAML(unmarshal func(interface{}) error) error {

	type taskType Task // Use new type to avoid recursion
	if err := unmarshal((*taskType)(t)); err != nil {
		return err
	}

	for name, opt := range t.Options {
		opt.Name = name
	}

	return nil
}

// Dependencies returns a list of options that are required explicitly.
// This does not include interpolations.
func (t *Task) Dependencies() []string {
	var options []string

	for _, opt := range t.Options {
		options = append(options, opt.Dependencies()...)
	}
	for _, run := range t.Run {
		options = append(options, run.When.Dependencies()...)
	}

	return options
}

// Execute runs the Run scripts in the task.
func (t *Task) Execute() error {
	for _, r := range t.Run {
		if err := t.run(r); err != nil {
			return err
		}
	}

	return nil
}

// run executes a Run struct.
func (t *Task) run(r *run.Run) error {

	if ok, err := t.shouldRun(r); !ok || err != nil {
		return err
	}

	if err := t.runCommands(r); err != nil {
		return err
	}

	if err := t.runSubTasks(r); err != nil {
		return err
	}

	if err := t.runEnvironment(r); err != nil {
		return err
	}

	return nil
}

func (t *Task) shouldRun(r *run.Run) (bool, error) {
	if r.When == nil {
		return true, nil
	}

	if err := r.When.Validate(t.Vars); err != nil {
		if !when.IsFailedCondition(err) {
			return false, err
		}

		for _, command := range r.Command {
			ui.PrintSkipped(command, err.Error())
		}

		for _, subTaskName := range r.Task {
			ui.PrintSkipped("task: "+subTaskName, err.Error())
		}

		return false, nil
	}

	return true, nil
}

func (t *Task) runCommands(r *run.Run) error {
	for _, command := range r.Command {
		if err := run.ExecCommand(command); err != nil {
			return err
		}
	}

	return nil
}

func (t *Task) runSubTasks(r *run.Run) error {
	for _, subTaskName := range r.Task {
		for _, subTask := range t.SubTasks {
			if subTask.Name == subTaskName {
				if err := subTask.Execute(); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (t *Task) runEnvironment(r *run.Run) error {
	ui.PrintEnvironment(r.Environment)
	for key, value := range r.Environment {
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
