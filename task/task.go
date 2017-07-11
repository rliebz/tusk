package task

// Task is a single task to be run by CLI
type Task struct {
	Args map[string]*Arg `yaml:",omitempty"`
	// TODO: Pre tasks should allow when conditionals.
	// Consider:
	//	type PreTask struct {
	//		When When // Factor out `When` struct for reusability
	//		Task string // Same purpose as `PreName`
	//	}
	PreName []string `yaml:"pre,omitempty"`
	Script  []Script
	Usage   string `yaml:",omitempty"`

	// Private members not specified in yaml file
	PreTasks []*Task `yaml:"-"`
}

// Execute runs the scripts in the task.
func (task *Task) Execute() error {
	// TODO: Announce task

	for _, preTask := range task.PreTasks {
		if err := preTask.Execute(); err != nil {
			return err
		}
		// fmt.Println("Task Finished!")
	}

	for _, script := range task.Script {
		if err := script.Execute(); err != nil {
			return err
		}
	}
	return nil
}
