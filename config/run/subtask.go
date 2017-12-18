package run

import "github.com/rliebz/tusk/config/marshal"

// SubTask is a description of a sub-task with passed options.
type SubTask struct {
	Name    string
	Options map[string]string
}

// UnmarshalYAML allows unmarshalling a string to represent the subtask name.
func (s *SubTask) UnmarshalYAML(unmarshal func(interface{}) error) error {

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

// SubTaskList is a list of subtasks with custom yaml unmarshalling.
type SubTaskList []*SubTask

// UnmarshalYAML allows single items to be used as lists.
func (l *SubTaskList) UnmarshalYAML(unmarshal func(interface{}) error) error { // nolint: dupl

	var subTaskSlice []*SubTask
	sliceCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&subTaskSlice) },
		Assign:    func() { *l = subTaskSlice },
	}

	var subTaskItem *SubTask
	itemCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&subTaskItem) },
		Assign:    func() { *l = SubTaskList{subTaskItem} },
	}

	return marshal.UnmarshalOneOf(sliceCandidate, itemCandidate)
}
