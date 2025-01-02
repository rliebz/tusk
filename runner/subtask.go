package runner

import "github.com/rliebz/tusk/marshal"

// SubTask is a description of a sub-task with passed options.
type SubTask struct {
	Name    string
	Args    marshal.Slice[string]
	Options map[string]string
}

// UnmarshalYAML allows unmarshaling a string to represent the subtask name.
func (s *SubTask) UnmarshalYAML(unmarshal func(any) error) error {
	var name string
	nameCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&name) },
		Assign:    func() { *s = SubTask{Name: name} },
	}

	type subTaskType SubTask // Use new type to avoid recursion
	var subTaskItem subTaskType
	subTaskCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&subTaskItem) },
		Assign:    func() { *s = SubTask(subTaskItem) },
	}

	return marshal.UnmarshalOneOf(nameCandidate, subTaskCandidate)
}
