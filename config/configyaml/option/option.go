package option

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/rliebz/tusk/config/configyaml/when"
)

// Option represents an abstract command line option.
type Option struct {
	Short   string
	Type    string
	Usage   string
	Private bool

	// Used to determine value
	Environment   string
	DefaultValues valueList `yaml:"default"`

	// Computed members not specified in yaml file
	Name   string `yaml:"-"`
	Passed string `yaml:"-"`
	Vars   map[string]string
}

// Dependencies returns a list of options that are required explicitly.
// This does not include interpolations.
func (o *Option) Dependencies() []string {
	var options []string

	for _, value := range o.DefaultValues {
		options = append(options, value.When.Dependencies()...)
	}

	return options
}

// Value determines an option's final value based on all configuration.
//
// For non-private variables, the order of priority is:
//   1. Parameter that was passed
//   2. Environment variable set
//   3. The first item in the default value list with a valid when clause
func (o *Option) Value() (string, error) {

	if o == nil {
		return "", nil
	}

	if !o.Private {
		if o.Passed != "" {
			return o.Passed, nil
		}

		envValue := os.Getenv(o.Environment)
		if envValue != "" {
			return envValue, nil
		}
	} else {
		if o.Environment != "" {
			return "", fmt.Errorf(
				`environment "%s" defined for private option`,
				o.Environment,
			)
		}
	}

	for _, candidate := range o.DefaultValues {
		if err := candidate.When.Validate(o.Vars); err != nil {
			continue
		}

		value, err := candidate.commandValueOrDefault()
		if err != nil {
			return "", errors.Wrapf(err, "could not compute value for flag: %s", o.Name)
		}

		return value, nil
	}

	return "", nil
}

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
	var valueItem *valueType
	if err = unmarshal(&valueItem); err == nil {
		*v = *(*value)(valueItem)

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
