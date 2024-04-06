package runner

import (
	"fmt"
	"strings"

	"github.com/rliebz/tusk/marshal"
)

// Value represents a value candidate for an option.
// When the when condition is true, either the command or value will be used.
type Value struct {
	When    WhenList
	Command string
	Value   string
}

// commandValueOrDefault validates a content definition, then gets the value.
func (v *Value) commandValueOrDefault(ctx Context) (string, error) {
	if v.Command != "" {
		cmd := newCmd(ctx, v.Command)

		out, err := cmd.Output()
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(string(out)), nil
	}

	return v.Value, nil
}

// UnmarshalYAML allows plain strings to represent a full struct. The value of
// the string is used as the Default field.
func (v *Value) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var valueString string
	stringCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&valueString) },
		Assign:    func() { *v = Value{Value: valueString} },
	}

	type valueType Value // Use new type to avoid recursion
	var valueItem valueType
	valueCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&valueItem) },
		Assign:    func() { *v = Value(valueItem) },
		Validate: func() error {
			if valueItem.Value != "" && valueItem.Command != "" {
				return fmt.Errorf(
					"value (%s) and command (%s) are both defined",
					valueItem.Value, valueItem.Command,
				)
			}

			return nil
		},
	}

	return marshal.UnmarshalOneOf(stringCandidate, valueCandidate)
}
