package config

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/rliebz/tusk/config/marshal"
)

// Value represents a value candidate for an option.
// When the when condition is true, either the command or value will be used.
type Value struct {
	When    List
	Command string
	Value   string
}

// commandValueOrDefault validates a content definition, then gets the value.
func (v *Value) commandValueOrDefault() (string, error) {
	if v.Command != "" {
		out, err := exec.Command("sh", "-c", v.Command).Output() // nolint: gosec
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

// ValueList is a slice of values with custom unmarshaling.
type ValueList []Value

// UnmarshalYAML allows single items to be used as lists.
func (vl *ValueList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var valueSlice []Value
	sliceCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&valueSlice) },
		Assign:    func() { *vl = valueSlice },
	}

	var valueItem Value
	itemCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&valueItem) },
		Assign:    func() { *vl = ValueList{valueItem} },
	}

	return marshal.UnmarshalOneOf(sliceCandidate, itemCandidate)
}

// ValueWithList is a list of allowable values for an option or argument.
type ValueWithList struct {
	ValuesAllowed marshal.StringList `yaml:"values"`
}

func (v *ValueWithList) validateSpecified(value, descriptor string) error {
	if len(v.ValuesAllowed) == 0 {
		return nil
	}

	for _, expected := range v.ValuesAllowed {
		if value == expected {
			return nil
		}
	}

	return fmt.Errorf(
		`value %q for %s must be one of %v`,
		value, descriptor, v.ValuesAllowed,
	)
}
