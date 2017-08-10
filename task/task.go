package task

// Task is a single task to be run by CLI
type Task struct {
	Args map[string]*Arg `yaml:",omitempty"`
	Pre  []struct {
		Name string
		When When
	} `yaml:",omitempty"`
	Script []Script
	Usage  string `yaml:",omitempty"`

	// Private members not specified in yaml file
	Name     string  `yaml:"-"`
	PreTasks []*Task `yaml:"-"`
}

// Execute runs the scripts in the task.
func (task *Task) Execute() error {
	// TODO: Announce task

	for _, preTask := range task.PreTasks {
		if err := preTask.Execute(); err != nil {
			return err
		}
	}

	for _, script := range task.Script {
		if err := script.Execute(); err != nil {
			return err
		}
	}
	return nil
}
