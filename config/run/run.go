package run

import (
	"github.com/rliebz/tusk/config/marshal"
	"github.com/rliebz/tusk/config/when"
)

// Run defines a a single runnable script within a task.
type Run struct {
	When    *when.When         `yaml:",omitempty"`
	Command marshal.StringList `yaml:",omitempty"`
	Task    marshal.StringList `yaml:",omitempty"`
}

// UnmarshalYAML allows plain strings to represent a run struct. The value of
// the string is used as the Command field.
func (r *Run) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var err error

	var command string
	if err = unmarshal(&command); err == nil {
		*r = Run{Command: marshal.StringList{command}}
		return nil
	}

	type runType Run // Use new type to avoid recursion
	var runItem *runType
	if err = unmarshal(&runItem); err == nil {
		*r = *(*Run)(runItem)
		return nil
	}

	return err
}

// List is a list of run items with custom yaml unmarshalling.
type List []*Run

// UnmarshalYAML allows single items to be used as lists.
func (rl *List) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var err error

	var runSlice []*Run
	if err = unmarshal(&runSlice); err == nil {
		*rl = runSlice
		return nil
	}

	var runItem *Run
	if err = unmarshal(&runItem); err == nil {
		*rl = List{runItem}
		return nil
	}

	return err
}
