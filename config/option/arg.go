package option

import "github.com/pkg/errors"

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
