package task

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gitlab.com/rliebz/tusk/appyaml"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Option represents an abstract command line option.
type Option struct {
	Short   string
	Type    string
	Usage   string
	Private bool

	// Used to determine value, in order of highest priority
	Environment string
	Computed    []struct {
		When    appyaml.When
		content `yaml:",inline"`
	}
	content `yaml:",inline"`

	// Computed members not specified in yaml file
	Name   string `yaml:"-"`
	Passed string `yaml:"-"`
}

type content struct {
	Command string
	Default string
}

// CreateCLIFlag converts an Option into a cli.Flag.
func CreateCLIFlag(opt *Option) (cli.Flag, error) {

	name := opt.Name
	if opt.Short != "" {
		name = fmt.Sprintf("%s, %s", name, opt.Short)
	}

	opt.Type = strings.ToLower(opt.Type)
	switch opt.Type {
	case "int", "integer":
		return cli.IntFlag{
			Name:  name,
			Usage: opt.Usage,
		}, nil
	case "float", "float64", "double":
		return cli.Float64Flag{
			Name:  name,
			Usage: opt.Usage,
		}, nil
	case "bool", "boolean":
		return cli.BoolFlag{
			Name:  name,
			Usage: opt.Usage,
		}, nil
	case "string", "":
		return cli.StringFlag{
			Name:  name,
			Usage: opt.Usage,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported flag type `%s`", opt.Type)
	}
}

// Value determines an option's final value based on all configuration.
//
// For non-private variables, the order of priority is:
//   1. Parameter that was passed
//   2. Environment variable set
//   3. The first item in the computed list with a valid when clause
//   4. The default, which is either a plain string or the output of a command
func (o *Option) Value() (string, error) {

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
				"environment `%s` defined for private option",
				o.Environment,
			)
		}
	}

	for _, candidate := range o.Computed {
		if err := candidate.When.Validate(); err != nil {
			continue
		}

		value, err := candidate.commandValueOrDefault()
		if err != nil {
			return "", errors.Wrapf(err, "could not compute value for flag: %s", o.Name)
		}

		return value, nil
	}

	value, err := o.commandValueOrDefault()
	if err != nil {
		return "", errors.Wrapf(err, "could not compute value for flag: %s", o.Name)
	}

	return value, nil
}

// commandValueOrDefault validates a content definition, then gets the value.
func (vg *content) commandValueOrDefault() (string, error) {

	if vg.Default != "" && vg.Command != "" {
		return "", fmt.Errorf(
			"default (%s) and command (%s) are both defined",
			vg.Default, vg.Command,
		)
	}

	if vg.Command != "" {
		out, err := exec.Command("sh", "-c", vg.Command).Output() // nolint: gas
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(string(out)), nil
	}

	return vg.Default, nil
}
