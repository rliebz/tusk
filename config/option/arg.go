package option

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/rliebz/tusk/config/marshal"
	yaml "gopkg.in/yaml.v2"
)

// Arg represents a command-line argument.
type Arg struct {
	valueWithList `yaml:",inline"`

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

// GetArgsWithOrder returns both the arg map and the ordered names.
func GetArgsWithOrder(ms yaml.MapSlice) (map[string]*Arg, []string, error) {
	args := make(map[string]*Arg, len(ms))
	assign := func(name string, text []byte) error {
		var arg *Arg
		if err := yaml.Unmarshal(text, &arg); err != nil {
			return err
		}

		if arg == nil {
			return fmt.Errorf("argument %q cannot be defined as null", name)
		}

		arg.Name = name
		args[name] = arg
		return nil
	}

	ordered, err := marshal.ParseOrderedMap(ms, assign)
	return args, ordered, err
}
