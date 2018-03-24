package option

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/rliebz/tusk/config/when"
)

// Option represents an abstract command line option.
type Option struct {
	ValueWithList `yaml:",inline"`

	Short    string
	Type     string
	Usage    string
	Private  bool
	Required bool

	// Used to determine value
	Environment   string
	DefaultValues ValueList `yaml:"default"`

	// Computed members not specified in yaml file
	Name       string `yaml:"-"`
	Passed     string `yaml:"-"`
	cacheValue string `yaml:"-"`
	isCacheSet bool   `yaml:"-"`
}

// Dependencies returns a list of options that are required explicitly.
// This does not include interpolations.
func (o *Option) Dependencies() []string {
	options := make([]string, 0, len(o.DefaultValues))
	for _, value := range o.DefaultValues {
		options = append(options, value.When.Dependencies()...)
	}

	return options
}

// UnmarshalYAML ensures that the option definition is valid.
func (o *Option) UnmarshalYAML(unmarshal func(interface{}) error) error {

	type optionType Option // Use new type to avoid recursion
	if err := unmarshal((*optionType)(o)); err != nil {
		return err
	}

	if len(o.Short) > 1 {
		return fmt.Errorf(
			`option short name "%s" cannot exceed one character`,
			o.Short,
		)
	}

	if o.Private {
		if o.Required {
			return errors.New("option cannot be both private and required")
		}

		if o.Environment != "" {
			return fmt.Errorf(
				`environment variable "%s" defined for private option`,
				o.Environment,
			)
		}

		if len(o.ValuesAllowed) != 0 {
			return errors.New("option cannot be private and specify values")
		}
	}

	if o.Required && len(o.DefaultValues) > 0 {
		return errors.New("default value defined for required option")
	}

	return nil
}

// Evaluate determines an option's value.
//
// The order of priority is:
//   1. Command-line option passed
//   2. Environment variable set
//   3. The first item in the default value list with a valid when clause
//
// Values may also be cached to avoid re-running commands.
func (o *Option) Evaluate(vars map[string]string) (string, error) {
	if o == nil {
		return "", nil
	}

	value, err := o.getValue(vars)
	if err != nil {
		return "", err
	}

	o.cache(value)

	return value, nil
}

func (o *Option) getValue(vars map[string]string) (string, error) {
	if o.isCacheSet {
		return o.cacheValue, nil
	}

	if !o.Private {
		if value, found := o.getSpecified(); found {
			if err := o.validateSpecified(value, "option "+o.Name); err != nil {
				return "", err
			}

			return value, nil
		}
	}

	if o.Required {
		return "", fmt.Errorf("no value passed for required option: %s", o.Name)
	}

	return o.getDefaultValue(vars)
}

func (o *Option) getSpecified() (value string, found bool) {

	if o.Passed != "" {
		return o.Passed, true
	}

	envValue := os.Getenv(o.Environment)
	if envValue != "" {
		return envValue, true
	}

	return "", false
}

func (o *Option) getDefaultValue(vars map[string]string) (string, error) {
	for _, candidate := range o.DefaultValues {
		if err := candidate.When.Validate(vars); err != nil {
			if !when.IsFailedCondition(err) {
				return "", err
			}
			continue
		}

		value, err := candidate.commandValueOrDefault()
		if err != nil {
			return "", errors.Wrapf(err, "could not compute value for option: %s", o.Name)
		}

		return value, nil
	}

	if o.isNumeric() {
		return "0", nil
	}

	if o.isBoolean() {
		return "false", nil
	}

	return "", nil
}

func (o *Option) cache(value string) {
	o.isCacheSet = true
	o.cacheValue = value
}

func (o *Option) isNumeric() bool {
	switch strings.ToLower(o.Type) {
	case "int", "integer", "float", "float64", "double":
		return true
	default:
		return false
	}
}

func (o *Option) isBoolean() bool {
	switch strings.ToLower(o.Type) {
	case "bool", "boolean":
		return true
	default:
		return false
	}
}
