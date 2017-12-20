package task

import (
	"github.com/rliebz/tusk/config/option"
)

// Task is a single task to be run by CLI.
type Task struct {
	Options     map[string]*option.Option `yaml:",omitempty"`
	RunList     RunList                   `yaml:"run"`
	Usage       string                    `yaml:",omitempty"`
	Description string                    `yaml:",omitempty"`
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

	for name, opt := range t.Options {
		opt.Name = name
	}

	return nil
}

// Dependencies returns a list of options that are required explicitly.
// This does not include interpolations.
func (t *Task) Dependencies() []string {
	options := make([]string, 0, len(t.Options)+len(t.RunList))

	for _, opt := range t.Options {
		options = append(options, opt.Dependencies()...)
	}
	for _, run := range t.RunList {
		options = append(options, run.When.Dependencies()...)
	}

	return options
}

// Execute runs the Run scripts in the task.
func (t *Task) Execute() error {
	for _, r := range t.RunList {
		if err := t.run(r); err != nil {
			return err
		}
	}

	return nil
}

// run executes a Run struct.
func (t *Task) run(r *Run) error {
	if ok, err := r.shouldRun(t.Vars); !ok || err != nil {
		return err
	}

	if err := r.runCommands(); err != nil {
		return err
	}

	if err := r.runSubTasks(); err != nil {
		return err
	}

	if err := r.runEnvironment(); err != nil {
		return err
	}

	return nil
}
