package option

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/rliebz/tusk/config/marshal"
	"github.com/rliebz/tusk/config/when"
)

type value struct {
	When    when.When
	Command string
	Value   string
}

// commandValueOrDefault validates a content definition, then gets the value.
func (v *value) commandValueOrDefault() (string, error) {

	if v.Command != "" {
		out, err := exec.Command("sh", "-c", v.Command).Output() // nolint: gas
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(string(out)), nil
	}

	return v.Value, nil
}

// UnmarshalYAML allows plain strings to represent a full struct. The value of
// the string is used as the Default field.
func (v *value) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var valueString string
	stringCandidate := marshal.Candidate{
		Unmarshal: func() error { return unmarshal(&valueString) },
		Assign:    func() { *v = value{Value: valueString} },
	}

	type valueType value // Use new type to avoid recursion
	var valueItem valueType
	valueCandidate := marshal.Candidate{
		Unmarshal: func() error { return unmarshal(&valueItem) },
		Assign:    func() { *v = value(valueItem) },
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

	return marshal.OneOf(stringCandidate, valueCandidate)
}

type valueList []value

// UnmarshalYAML allows single items to be used as lists.
func (vl *valueList) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var valueSlice []value
	sliceCandidate := marshal.Candidate{
		Unmarshal: func() error { return unmarshal(&valueSlice) },
		Assign:    func() { *vl = valueSlice },
	}

	var valueItem value
	itemCandidate := marshal.Candidate{
		Unmarshal: func() error { return unmarshal(&valueItem) },
		Assign:    func() { *vl = valueList{valueItem} },
	}

	return marshal.OneOf(sliceCandidate, itemCandidate)
}
