package run

import (
	"errors"

	"github.com/rliebz/tusk/config/marshal"
	"github.com/rliebz/tusk/config/when"
)

// Run defines a a single runnable item within a task.
type Run struct {
	When        when.When          `yaml:",omitempty"`
	Command     marshal.StringList `yaml:",omitempty"`
	Task        SubTaskList        `yaml:",omitempty"`
	Environment map[string]*string `yaml:",omitempty"`
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
				len(runItem.Task) != 0,
				runItem.Environment != nil,
			}

			count := 0
			for _, isUsed := range actionUsedList {
				if isUsed {
					count++
				}
			}

			if count > 1 {
				return errors.New("Only one action can be defined in `run`")
			}

			return nil
		},
	}

	return marshal.UnmarshalOneOf(commandCandidate, runCandidate)
}

// List is a list of run items with custom yaml unmarshalling.
type List []*Run

// UnmarshalYAML allows single items to be used as lists.
func (rl *List) UnmarshalYAML(unmarshal func(interface{}) error) error { // nolint: dupl

	var runSlice []*Run
	sliceCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&runSlice) },
		Assign:    func() { *rl = runSlice },
	}

	var runItem *Run
	itemCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&runItem) },
		Assign:    func() { *rl = List{runItem} },
	}

	return marshal.UnmarshalOneOf(sliceCandidate, itemCandidate)
}
