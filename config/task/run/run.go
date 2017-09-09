package run

import (
	"github.com/pkg/errors"
	"github.com/rliebz/tusk/config/task/appyaml"
	"github.com/rliebz/tusk/config/task/when"
)

// Run defines a a single runnable script within a task.
type Run struct {
	When    *when.When         `yaml:",omitempty"`
	Command appyaml.StringList `yaml:",omitempty"`
	Task    appyaml.StringList `yaml:",omitempty"`
}

// UnmarshalYAML allows plain strings to represent a run struct. The value of
// the string is used as the Command field.
func (r *Run) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var command string
	if err := unmarshal(&command); err == nil {
		*r = Run{Command: appyaml.StringList{command}}
		return nil
	}

	type runType Run // Use new type to avoid recursion
	var runItem *runType
	if err := unmarshal(&runItem); err == nil {
		*r = *(*Run)(runItem)
		return nil
	}

	return errors.New("could not parse run item")
}

// List is a list of run items with custom yaml unmarshalling.
type List []*Run

// UnmarshalYAML allows single items to be used as lists.
func (rl *List) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var runItem *Run
	if err := unmarshal(&runItem); err == nil {
		*rl = List{runItem}
		return nil
	}

	var runSlice []*Run
	if err := unmarshal(&runSlice); err == nil {
		*rl = runSlice
		return nil
	}

	return errors.New("could not parse runlist")
}
