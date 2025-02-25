package runner

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/rliebz/tusk/marshal"
)

// Passable is a list of allowable values for an option or argument.
type Passable struct {
	Usage         string                `yaml:"usage"`
	Type          string                `yaml:"type"`
	ValuesAllowed marshal.Slice[string] `yaml:"values"`

	// Computed members not specified in yaml file
	Name   string `yaml:"-"`
	Passed string `yaml:"-"`
}

// validatePassed validates that the specified value is compatible with the
// passable configuration.
//
// The value should be the actual value passed. The kind should be the kind of
// passable, such as "option" or "argument".
func (p *Passable) validatePassed(kind string, value string) error {
	if len(p.ValuesAllowed) != 0 && !slices.Contains(p.ValuesAllowed, value) {
		return fmt.Errorf(
			`value %q for %s %q must be one of [%s]`,
			value, kind, p.Name, strings.Join(p.ValuesAllowed, ", "),
		)
	}

	if !p.hasValidType(value) {
		return fmt.Errorf(
			`value %q for %s %q is not of type %q`,
			value, kind, p.Name, p.Type,
		)
	}

	return nil
}

func (p *Passable) hasValidType(value string) bool {
	switch {
	case p.isBoolean():
		_, err := strconv.ParseBool(value)
		return err == nil
	case p.isInt():
		_, err := strconv.Atoi(value)
		return err == nil
	case p.isFloat():
		_, err := strconv.ParseFloat(value, 64)
		return err == nil
	}

	return true
}

func (p *Passable) isNumeric() bool {
	return p.isInt() || p.isFloat()
}

func (p *Passable) isFloat() bool {
	switch strings.ToLower(p.Type) {
	case "float", "float64", "double":
		return true
	default:
		return false
	}
}

func (p *Passable) isInt() bool {
	switch strings.ToLower(p.Type) {
	case "int", "integer":
		return true
	default:
		return false
	}
}

func (p *Passable) isBoolean() bool {
	switch strings.ToLower(p.Type) {
	case "bool", "boolean":
		return true
	default:
		return false
	}
}
