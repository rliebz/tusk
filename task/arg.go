package task

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	// `Default` and `Computed` are mutually exclusive.
	Passed      string `yaml:"-"`
	Environment string
	Default     interface{} // TODO: This can probably be a string
	Computed    string

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
	return cli.IntFlag{
		Name:  name,
		Usage: arg.Usage,
	}, nil
}

func createFloatFlag(name string, arg *Arg) (cli.Flag, error) {
	return cli.Float64Flag{
		Name:  name,
		Usage: arg.Usage,
	}, nil
}

func createBoolFlag(name string, arg *Arg) (cli.Flag, error) {
	return cli.BoolFlag{
		Name:  name,
		Usage: arg.Usage,
	}, nil
}

func createStringFlag(name string, arg *Arg) (cli.Flag, error) {
	return cli.StringFlag{
		Name:  name,
		Usage: arg.Usage,
	}, nil
}

// Value determines the final argument value based on all options.
func (arg *Arg) Value() (string, error) {
	if arg.Default != nil && arg.Computed != "" {
		return "", fmt.Errorf(
			"default and computed are both defined for flag: %v",
			arg.Name,
		)
	}

	if arg.Passed != "" {
		return arg.Passed, nil
	}

	envValue := os.Getenv(arg.Environment)
	if envValue != "" {
		return envValue, nil
	}

	if arg.Default != nil {
		return fmt.Sprint(arg.Default), nil
	}

	out, err := exec.Command("sh", "-c", arg.Computed).Output() // nolint: gas
	if err != nil {
		return "", errors.Wrapf(err, "could not compute value for %s", arg.Name)
	}

	return strings.TrimSpace(string(out)), nil
}
