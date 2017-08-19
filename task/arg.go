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

// Arg represents an abstract command line argument.
type Arg struct {
	Short   string
	Type    string
	Usage   string
	Private bool

	// Used to determine value, in order of highest priority
	// `Command` and `Default` are mutually exclusive
	Passed      string `yaml:"-"`
	Environment string
	Computed    []computed
	Command     string
	Default     string

	// Computed members not specified in yaml file
	Name string `yaml:"-"`
}

func (a *Arg) getCommand() string { return a.Command }
func (a *Arg) getDefault() string { return a.Default }

type computed struct {
	When    appyaml.When
	Command string
	Default string
}

func (c *computed) getCommand() string { return c.Command }
func (c *computed) getDefault() string { return c.Default }

// CreateCLIFlag converts an Arg into a cli.Flag.
func CreateCLIFlag(arg *Arg) (cli.Flag, error) {

	name := arg.Name
	if arg.Short != "" {
		name = fmt.Sprintf("%s, %s", name, arg.Short)
	}

	arg.Type = strings.ToLower(arg.Type)
	switch arg.Type {
	case "int", "integer":
		return cli.IntFlag{
			Name:  name,
			Usage: arg.Usage,
		}, nil
	case "float", "float64", "double":
		return cli.Float64Flag{
			Name:  name,
			Usage: arg.Usage,
		}, nil
	case "bool", "boolean":
		return cli.BoolFlag{
			Name:  name,
			Usage: arg.Usage,
		}, nil
	case "string", "":
		return cli.StringFlag{
			Name:  name,
			Usage: arg.Usage,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported flag type `%s`", arg.Type)
	}
}

// Value determines the final argument value based on all options.
func (a *Arg) Value() (string, error) {

	if !a.Private {
		if a.Passed != "" {
			return a.Passed, nil
		}

		envValue := os.Getenv(a.Environment)
		if envValue != "" {
			return envValue, nil
		}
	}

	for _, candidate := range a.Computed {
		if err := candidate.When.Validate(); err != nil {
			continue
		}

		value, err := getCommandOrDefault(&candidate)
		if err != nil {
			return "", errors.Wrapf(err, "could not compute value for flag: %s", a.Name)
		}

		return value, nil
	}

	value, err := getCommandOrDefault(a)
	if err != nil {
		return "", errors.Wrapf(err, "could not compute value for flag: %s", a.Name)
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
