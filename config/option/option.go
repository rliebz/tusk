package option

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/rliebz/tusk/config/when"
)

// Option represents an abstract command line option.
type Option struct {
	Short    string
	Type     string
	Usage    string
	Private  bool
	Required bool

	// Used to determine value
	Environment   string
	DefaultValues valueList `yaml:"default"`

	// Computed members not specified in yaml file
	Name       string            `yaml:"-"`
	Passed     string            `yaml:"-"`
	Vars       map[string]string `yaml:"-"`
	cacheValue string            `yaml:"-"`
	isCacheSet bool              `yaml:"-"`
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

	if o.Private && o.Required {
		return errors.New("option cannot be both private and required")
	}

	if o.Private && o.Environment != "" {
		return fmt.Errorf(
			`environment variable "%s" defined for private option`,
			o.Environment,
		)
	}

	if o.Required && len(o.DefaultValues) > 0 {
		return errors.New("default value defined for required option")
	}

	return nil
}

// Value determines an option's final value based on all configuration.
//
// The order of priority is:
//   1. Command-line option passed
//   2. Environment variable set
//   3. The first item in the default value list with a valid when clause
//
// Values may also be cached to avoid re-running commands.
func (o *Option) Value() (string, error) {

	if o == nil {
		return "", nil
	}

	if o.isCacheSet {
		return o.cacheValue, nil
	}

	if !o.Private {
		if o.Passed != "" {
			return o.Passed, nil
		}

		envValue := os.Getenv(o.Environment)
		if envValue != "" {
			return envValue, nil
		}
	}

	if o.Required {
		return "", fmt.Errorf("no value passed for required option: %s", o.Name)
	}

	return o.getDefaultValue()
}

func (o *Option) getDefaultValue() (string, error) {
	for _, candidate := range o.DefaultValues {
		if err := candidate.When.Validate(o.Vars); err != nil {
			if !when.IsFailedCondition(err) {
				return "", err
			}
			continue
		}

		value, err := candidate.commandValueOrDefault()
		if err != nil {
			return "", errors.Wrapf(err, "could not compute value for option: %s", o.Name)
		}

		o.cache(value)
		return value, nil
	}

	return "", nil
}

func (o *Option) cache(value string) {
	o.isCacheSet = true
	o.cacheValue = value
}
