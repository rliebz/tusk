package runner

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/rliebz/tusk/marshal"
	yaml "gopkg.in/yaml.v2"
)

// Option represents an abstract command line option.
type Option struct {
	Passable `yaml:",inline"`

	Short    string
	Private  bool
	Required bool
	Rewrite  string

	// Used to determine value
	Environment   string
	DefaultValues ValueList `yaml:"default"`

	// Computed members not specified in yaml file
	cacheValue string `yaml:"-"`
	isCacheSet bool   `yaml:"-"`
}

// Equal provides a method of checking option equality for testing purposes only.
func (o *Option) Equal(other *Option) bool {
	return reflect.DeepEqual(o, other)
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

	return o.validate()
}

func (o *Option) validate() error {
	if len(o.Short) > 1 {
		return fmt.Errorf(
			"option short name %q cannot exceed one character",
			o.Short,
		)
	}

	if o.Private {
		if err := o.validatePrivate(); err != nil {
			return err
		}
	}

	if o.Required && len(o.DefaultValues) > 0 {
		return errors.New("default value defined for required option")
	}

	if o.Rewrite != "" && !o.isBoolean() {
		return errors.New("rewrite may only be performed on boolean values")
	}

	return nil
}

func (o *Option) validatePrivate() error {
	if o.Required {
		return errors.New("option cannot be both private and required")
	}

	if o.Environment != "" {
		return fmt.Errorf(
			"environment variable %q defined for private option",
			o.Environment,
		)
	}

	if o.Short != "" {
		return errors.New("option cannot be private and specify a short name")
	}

	if len(o.ValuesAllowed) != 0 {
		return errors.New("option cannot be private and specify values")
	}

	return nil
}

func (o *Option) validatePassed(value string) error {
	return o.Passable.validatePassed("option", value)
}

// Evaluate determines an option's value.
//
// The order of priority is:
//  1. Command-line option passed
//  2. Environment variable set
//  3. The first item in the default value list with a valid when clause
//
// Values may also be cached to avoid re-running commands.
func (o *Option) Evaluate(ctx Context, vars map[string]string) (string, error) {
	if o == nil {
		return "", nil
	}

	value, err := o.getValue(ctx, vars)
	if err != nil {
		return "", err
	}

	o.cache(value)

	return value, nil
}

func (o *Option) getValue(ctx Context, vars map[string]string) (string, error) {
	if o.isCacheSet {
		return o.cacheValue, nil
	}

	if !o.Private {
		if value, found := o.getSpecified(); found {
			if err := o.validatePassed(value); err != nil {
				return "", err
			}

			return value, nil
		}
	}

	if o.Required {
		return "", fmt.Errorf("no value passed for required option: %s", o.Name)
	}

	return o.getDefaultValue(ctx, vars)
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

func (o *Option) getDefaultValue(ctx Context, vars map[string]string) (string, error) {
	for _, candidate := range o.DefaultValues {
		if err := candidate.When.Validate(ctx, vars); err != nil {
			if !IsFailedCondition(err) {
				return "", err
			}
			continue
		}

		value, err := candidate.commandValueOrDefault(ctx)
		if err != nil {
			return "", fmt.Errorf("could not compute value for option %q: %w", o.Name, err)
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

// Options represents an ordered set of options as specified in the config.
type Options []*Option

// UnmarshalYAML unmarshals an ordered set of options and assigns names.
func (o *Options) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ms yaml.MapSlice
	if err := unmarshal(&ms); err != nil {
		return err
	}

	options, err := getOptionsWithOrder(ms)
	if err != nil {
		return err
	}

	*o = options

	return nil
}

// Lookup finds an Option by name.
func (o *Options) Lookup(name string) (*Option, bool) {
	for _, opt := range *o {
		if opt.Name == name {
			return opt, true
		}
	}

	return nil, false
}

// getOptionsWithOrder returns both the option map and the ordered names.
func getOptionsWithOrder(ms yaml.MapSlice) ([]*Option, error) {
	options := make([]*Option, 0, len(ms))
	assign := func(name string, text []byte) error {
		var opt Option
		if err := yaml.UnmarshalStrict(text, &opt); err != nil {
			return err
		}
		opt.Name = name

		options = append(options, &opt)

		return nil
	}

	_, err := marshal.ParseOrderedMap(ms, assign)
	return options, err
}
