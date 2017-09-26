package option

import (
	"fmt"
	"os/exec"
	"strings"

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

	var err error

	var valueString string
	if err = unmarshal(&valueString); err == nil {
		*v = value{Value: valueString}
		return nil
	}

	type valueType value // Use new type to avoid recursion
	if err = unmarshal((*valueType)(v)); err == nil {

		if v.Value != "" && v.Command != "" {
			return fmt.Errorf(
				"value (%s) and command (%s) are both defined",
				v.Value, v.Command,
			)
		}

		return nil
	}

	return err
}

type valueList []value

// UnmarshalYAML allows single items to be used as lists.
func (vl *valueList) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var err error

	var valueSlice []value
	if err = unmarshal(&valueSlice); err == nil {
		*vl = valueSlice
		return nil
	}

	var valueItem value
	if err = unmarshal(&valueItem); err == nil {
		*vl = valueList{valueItem}
		return nil
	}

	return err
}
