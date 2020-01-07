package config

import (
	"github.com/pkg/errors"
	"github.com/rliebz/tusk/config/marshal"
	yaml "gopkg.in/yaml.v2"
)

// Arg represents a command-line argument.
type Arg struct {
	ValueWithList `yaml:",inline"`

	Usage string

	// Computed members not specified in yaml file
	Name   string `yaml:"-"`
	Passed string `yaml:"-"`
}

// Evaluate determines an argument's value.
func (a *Arg) Evaluate() (string, error) {
	if a == nil {
		return "", errors.New("nil argument evaluated")
	}

	if err := a.validateSpecified(a.Passed, "argument "+a.Name); err != nil {
		return "", err
	}

	return a.Passed, nil
}

// Args represents an ordered set of arguments as specified in the config.
type Args []*Arg

// UnmarshalYAML unmarshals an ordered set of options and assigns names.
func (a *Args) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ms yaml.MapSlice
	if err := unmarshal(&ms); err != nil {
		return err
	}

	args, err := getArgsWithOrder(ms)
	if err != nil {
		return err
	}

	*a = args

	return nil
}

// Lookup finds an Arg by name.
func (a *Args) Lookup(name string) (*Arg, bool) {
	for _, arg := range *a {
		if arg.Name == name {
			return arg, true
		}
	}

	return nil, false
}

// getArgsWithOrder returns both the arg map and the ordered names.
func getArgsWithOrder(ms yaml.MapSlice) ([]*Arg, error) {
	args := make([]*Arg, 0, len(ms))
	assign := func(name string, text []byte) error {
		var arg Arg
		if err := yaml.UnmarshalStrict(text, &arg); err != nil {
			return err
		}

		arg.Name = name

		args = append(args, &arg)

		return nil
	}

	_, err := marshal.ParseOrderedMap(ms, assign)
	return args, err
}
