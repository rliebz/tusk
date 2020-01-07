package runner

import (
	"errors"

	"github.com/rliebz/tusk/marshal"
	"github.com/rliebz/tusk/ui"
)

// Run defines a a single runnable item within a task.
type Run struct {
	When           WhenList           `yaml:",omitempty"`
	Command        CommandList        `yaml:",omitempty"`
	SubTaskList    SubTaskList        `yaml:"task,omitempty"`
	SetEnvironment map[string]*string `yaml:"set-environment,omitempty"`

	// Computed members not specified in yaml file
	Tasks []Task `yaml:"-"`
}

// UnmarshalYAML allows simple commands to represent run structs.
func (r *Run) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cl CommandList
	commandCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&cl) },
		Assign:    func() { *r = Run{Command: cl} },
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

			return nil
		},
	}

	return marshal.UnmarshalOneOf(commandCandidate, runCandidate)
}

func (r *Run) shouldRun(vars map[string]string) (bool, error) {
	if err := r.When.Validate(vars); err != nil {
		if !IsFailedCondition(err) {
			return false, err
		}

		for _, command := range r.Command {
			ui.PrintSkipped(command.Print, err.Error())
		}

		for _, subTask := range r.SubTaskList {
			ui.PrintSkipped("task: "+subTask.Name, err.Error())
		}

		return false, nil
	}

	return true, nil
}

// RunList is a list of run items with custom yaml unmarshaling.
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
