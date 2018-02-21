package task

import (
	"errors"
	"os"

	"github.com/rliebz/tusk/config/marshal"
	"github.com/rliebz/tusk/config/when"
	"github.com/rliebz/tusk/ui"
)

// Run defines a a single runnable item within a task.
type Run struct {
	When           when.List          `yaml:",omitempty"`
	Command        marshal.StringList `yaml:",omitempty"`
	SubTaskList    SubTaskList        `yaml:"task,omitempty"`
	SetEnvironment map[string]*string `yaml:"set-environment,omitempty"`
	// Deprecated: Use SetEnvironment instead
	Environment map[string]*string `yaml:",omitempty"`

	// Computed members not specified in yaml file
	Tasks []Task `yaml:"-"`
}

// UnmarshalYAML allows plain strings to represent a run struct. The value of
// the string is used as the Command field.
func (r *Run) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var command string
	commandCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&command) },
		Assign:    func() { *r = Run{Command: marshal.StringList{command}} },
	}

	type runType Run // Use new type to avoid recursion
	var runItem runType
	runCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&runItem) },
		Assign:    func() { *r = Run(runItem) },
		Validate: func() error {
			actionUsedList := []bool{
				len(runItem.Command) != 0,
				len(runItem.SubTaskList) != 0,
				runItem.Environment != nil,
				runItem.SetEnvironment != nil,
			}

			count := 0
			for _, isUsed := range actionUsedList {
				if isUsed {
					count++
				}
			}

			if count > 1 {
				return errors.New("only one action can be defined in `run`")
			}

			if len(runItem.Environment) > 0 {
				ui.Deprecate(
					"The `environment` key has been deprecated in `run` clauses",
					"Use `set-environment` instead\n",
				)
				runItem.SetEnvironment = runItem.Environment
				runItem.Environment = nil
			}

			return nil
		},
	}

	return marshal.UnmarshalOneOf(commandCandidate, runCandidate)
}

func (r *Run) shouldRun(vars map[string]string) (bool, error) {
	if err := r.When.Validate(vars); err != nil {
		if !when.IsFailedCondition(err) {
			return false, err
		}

		for _, command := range r.Command {
			ui.PrintSkipped(command, err.Error())
		}

		for _, subTask := range r.SubTaskList {
			ui.PrintSkipped("task: "+subTask.Name, err.Error())
		}

		return false, nil
	}

	return true, nil
}

func (r *Run) runCommands() error {
	for _, command := range r.Command {
		if err := execCommand(command); err != nil {
			return err
		}
	}

	return nil
}

func (r *Run) runSubTasks() error {
	for _, subTask := range r.Tasks {
		if err := subTask.Execute(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Run) runEnvironment() error {
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

// RunList is a list of run items with custom yaml unmarshalling.
type RunList []*Run

// UnmarshalYAML allows single items to be used as lists.
func (rl *RunList) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var runSlice []*Run
	sliceCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&runSlice) },
		Assign:    func() { *rl = runSlice },
	}

	var runItem *Run
	itemCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&runItem) },
		Assign:    func() { *rl = RunList{runItem} },
	}

	return marshal.UnmarshalOneOf(sliceCandidate, itemCandidate)
}
