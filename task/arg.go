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
	// `Command` and `Default` are mutually exclusive
	Environment string
	Computed    []computed
	Command     string
	Default     string

	// Computed members not specified in yaml file
	Name   string `yaml:"-"`
	Passed string `yaml:"-"`
}

func (o *Option) getCommand() string { return o.Command }
func (o *Option) getDefault() string { return o.Default }

type computed struct {
	When    appyaml.When
	Command string
	Default string
}

func (c *computed) getCommand() string { return c.Command }
func (c *computed) getDefault() string { return c.Default }

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

// Value determines the final argument value based on all options.
func (o *Option) Value() (string, error) {

	if !o.Private {
		if o.Passed != "" {
			return o.Passed, nil
		}

		envValue := os.Getenv(o.Environment)
		if envValue != "" {
			return envValue, nil
		}
	}

	for _, candidate := range o.Computed {
		if err := candidate.When.Validate(); err != nil {
			continue
		}

		value, err := getCommandOrDefault(&candidate)
		if err != nil {
			return "", errors.Wrapf(err, "could not compute value for flag: %s", o.Name)
		}

		return value, nil
	}

	value, err := getCommandOrDefault(o)
	if err != nil {
		return "", errors.Wrapf(err, "could not compute value for flag: %s", o.Name)
	}

	return value, nil
}

// valueGetter determines a value by running a command or using a default.
// While both options are available, it is required that only one of the two
// are defined, since neither has an innate priority over the other.
type valueGetter interface {
	getCommand() string
	getDefault() string
}

func getCommandOrDefault(vg valueGetter) (string, error) {

	if vg.getDefault() != "" && vg.getCommand() != "" {
		return "", fmt.Errorf(
			"default (%s) and command (%s) are both defined",
			vg.getDefault(), vg.getCommand(),
		)
	}

	if vg.getCommand() != "" {
		out, err := exec.Command("sh", "-c", vg.getCommand()).Output() // nolint: gas
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(string(out)), nil
	}

	return vg.getDefault(), nil
}
