package task

import (
	"fmt"

	"github.com/rliebz/tusk/config/configyaml/option"
	"github.com/rliebz/tusk/config/configyaml/run"
	"github.com/rliebz/tusk/ui"
)

// Task is a single task to be run by CLI.
type Task struct {
	Options     map[string]*option.Option `yaml:",omitempty"`
	Run         run.List
	Usage       string `yaml:",omitempty"`
	Description string `yaml:",omitempty"`

	// Computed members not specified in yaml file
	Name     string  `yaml:"-"`
	SubTasks []*Task `yaml:"-"`
	Vars     map[string]string
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
	// TODO: Announce task

	for _, r := range t.Run {
		if err := t.run(r); err != nil {
			return err
		}
	}

	return nil
}

// run executes a Run struct.
func (t *Task) run(r *run.Run) error {

	// TODO: Validation logic should happen before runtime.
	if err := t.validateRun(r); err != nil {
		return err
	}

	if ok := t.shouldRun(r); !ok {
		return nil
	}

	if err := t.runCommands(r); err != nil {
		return err
	}

	if err := t.runSubTasks(r); err != nil {
		return err
	}

	return nil
}

func (t *Task) validateRun(r *run.Run) error {
	if len(r.Command) != 0 && len(r.Task) != 0 {
		return fmt.Errorf(
			"subtask (%s) and command (%s) are both defined",
			r.Command, r.Task,
		)
	}

	return nil
}

func (t *Task) shouldRun(r *run.Run) (ok bool) {

	if r.When == nil {
		return true
	}

	if err := r.When.Validate(t.Vars); err != nil {
		for _, command := range r.Command {
			ui.PrintSkipped(command, err.Error())
		}
		for _, subTaskName := range r.Task {
			ui.PrintSkipped("task: "+subTaskName, err.Error())
		}
		return false
	}

	return true
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
