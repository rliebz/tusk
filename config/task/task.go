package task

import (
	"fmt"

	"github.com/rliebz/tusk/ui"
)

// Task is a single task to be run by CLI.
type Task struct {
	Options     map[string]*Option `yaml:",omitempty"`
	Run         []*Run
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

	for _, run := range t.Run {
		if err := t.run(run); err != nil {
			return err
		}
	}

	return nil
}

// run executes a Run struct.
func (t *Task) run(run *Run) error {

	// TODO: Validation logic should happen before runtime.
	if err := t.validateRun(run); err != nil {
		return err
	}

	if ok := t.shouldRun(run); !ok {
		return nil
	}

	if err := t.runCommands(run); err != nil {
		return err
	}

	if err := t.runSubTasks(run); err != nil {
		return err
	}

	return nil
}

func (t *Task) validateRun(run *Run) error {
	if len(run.Command.Values) != 0 && len(run.Task.Values) != 0 {
		return fmt.Errorf(
			"subtask (%s) and command (%s) are both defined",
			run.Command.Values, run.Task.Values,
		)
	}

	return nil
}

func (t *Task) shouldRun(run *Run) (ok bool) {

	if run.When == nil {
		return true
	}

	if err := run.When.Validate(t.Vars); err != nil {
		for _, command := range run.Command.Values {
			ui.PrintSkipped(command, err.Error())
		}
		for _, subTaskName := range run.Task.Values {
			ui.PrintSkipped("task: "+subTaskName, err.Error())
		}
		return false
	}

	return true
}

func (t *Task) runCommands(run *Run) error {
	for _, command := range run.Command.Values {
		if err := execCommand(command); err != nil {
			return err
		}
	}

	return nil
}

func (t *Task) runSubTasks(run *Run) error {
	for _, subTaskName := range run.Task.Values {
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
