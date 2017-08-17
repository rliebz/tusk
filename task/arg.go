package task

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
)

// Arg represents an abstract command line argument.
type Arg struct {
	Short   string
	Type    string
	Usage   string
	Private bool

	// Used to determine value, in order of highest priority
	Passed      string `yaml:"-"`
	Environment string
	Computed    []struct {
		When  When
		Value string
	}
	Default string

	// Computed members not specified in yaml file
	Name string `yaml:"-"`
}

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
func (arg *Arg) Value() (string, error) {

	if !arg.Private {
		if arg.Passed != "" {
			return arg.Passed, nil
		}

		envValue := os.Getenv(arg.Environment)
		if envValue != "" {
			return envValue, nil
		}
	}

	for _, candidate := range arg.Computed {
		if err := candidate.When.Validate(); err != nil {
			continue
		}
		return candidate.Value, nil
	}

	return arg.Default, nil
}
