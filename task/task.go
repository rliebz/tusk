package task

import (
	"gitlab.com/rliebz/tusk/ui"
)

// Task is a single task to be run by CLI
type Task struct {
	Args map[string]*Arg `yaml:",omitempty"`
	Pre  []struct {
		Name string
		When When
	} `yaml:",omitempty"`
	Script []Script
	Usage  string `yaml:",omitempty"`

	// Computed members not specified in yaml file
	Name     string  `yaml:"-"`
	PreTasks []*Task `yaml:"-"`
}

// Execute runs the scripts in the task.
func (t *Task) Execute() error {
	// TODO: Announce task

	for _, preTask := range t.PreTasks {

		var when When
		for _, p := range t.Pre {
			if p.Name == preTask.Name {
				when = p.When
				break
			}
		}

		if err := when.Validate(); err != nil {
			ui.PrintCommandSkipped("pre-task: "+preTask.Name, err.Error())
			continue
		}

		if err := preTask.Execute(); err != nil {
			return err
		}
	}

	for _, script := range t.Script {
		if err := script.Execute(); err != nil {
			return err
		}
	}
	return nil
}
