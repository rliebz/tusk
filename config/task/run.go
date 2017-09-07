package task

import (
	"github.com/pkg/errors"
	"github.com/rliebz/tusk/appyaml"
)

// run defines a a single runnable script within a task.
type run struct {
	When    *appyaml.When      `yaml:",omitempty"`
	Command appyaml.StringList `yaml:",omitempty"`
	Task    appyaml.StringList `yaml:",omitempty"`
}

// UnmarshalYAML allows plain strings to represent a run struct. The value of
// the string is used as the Command field.
func (r *run) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var command string
	if err := unmarshal(&command); err == nil {
		*r = run{Command: appyaml.StringList{command}}
		return nil
	}

	type runType run // Use new type to avoid recursion
	var runItem *runType
	if err := unmarshal(&runItem); err == nil {
		*r = *(*run)(runItem)
		return nil
	}

	return errors.New("could not parse run item")
}

type runList []*run

// UnmarshalYAML allows single items to be used as lists.
func (rl *runList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var runItem *run
	if err := unmarshal(&runItem); err == nil {
		*rl = runList{runItem}
		return nil
	}

	var runSlice []*run
	if err := unmarshal(&runSlice); err == nil {
		*rl = runSlice
		return nil
	}

	return errors.New("could not parse runlist")
}
