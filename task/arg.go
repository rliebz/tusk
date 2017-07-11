package task

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

// Arg represents an abstract command line argument.
type Arg struct {
	Short       string
	Default     interface{}
	Environment string
	Type        string
	Usage       string

	// Private members not specified in yaml file
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
		return createIntFlag(name, arg)
	case "float", "float64", "double":
		return createFloatFlag(name, arg)
	case "bool", "boolean":
		return createBoolFlag(name, arg)
	case "string", "":
		return createStringFlag(name, arg)
	default:
		return nil, fmt.Errorf("unsupported flag type `%s`", arg.Type)
	}
}

func createIntFlag(name string, arg *Arg) (cli.Flag, error) {
	value, ok := arg.Default.(int)
	if arg.Default != nil && !ok {
		return nil, fmt.Errorf(
			"default value `%s` for arg `%s` is not of type int",
			arg.Default, name,
		)
	}

	return cli.IntFlag{
		Name:   name,
		Value:  value,
		Usage:  arg.Usage,
		EnvVar: arg.Environment,
	}, nil
}

func createFloatFlag(name string, arg *Arg) (cli.Flag, error) {
	value, ok := arg.Default.(float64)
	if arg.Default != nil && !ok {
		return nil, fmt.Errorf(
			"default value `%s` for arg `%s` is not of type float",
			arg.Default, name,
		)
	}

	return cli.Float64Flag{
		Name:   name,
		Value:  value,
		Usage:  arg.Usage,
		EnvVar: arg.Environment,
	}, nil
}

func createBoolFlag(name string, arg *Arg) (cli.Flag, error) {
	trueByDefault, ok := arg.Default.(bool)
	if arg.Default != nil && !ok {
		return nil, fmt.Errorf(
			"default value `%s` for arg `%s` is not of type bool",
			arg.Default, name,
		)
	}

	if trueByDefault {
		return cli.BoolTFlag{
			Name:   name,
			Usage:  arg.Usage,
			EnvVar: arg.Environment,
		}, nil
	}

	return cli.BoolFlag{
		Name:   name,
		Usage:  arg.Usage,
		EnvVar: arg.Environment,
	}, nil
}

func createStringFlag(name string, arg *Arg) (cli.Flag, error) {
	value, ok := arg.Default.(string)
	if arg.Default != nil && !ok {
		return nil, fmt.Errorf(
			"default value `%s` for arg `%s` is not of type string",
			arg.Default, name,
		)
	}

	return cli.StringFlag{
		Name:   name,
		Value:  value,
		Usage:  arg.Usage,
		EnvVar: arg.Environment,
	}, nil
}
