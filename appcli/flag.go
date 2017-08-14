package appcli

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	yaml "gopkg.in/yaml.v2"

	"gitlab.com/rliebz/tusk/config"
	"gitlab.com/rliebz/tusk/interp"
	"gitlab.com/rliebz/tusk/task"
)

// copyFlags copies all command flags from one cli.App to another.
func copyFlags(target *cli.App, source *cli.App) {
	for i, targetCommand := range target.Commands {
		for _, sourceCommand := range source.Commands {
			if targetCommand.Name == sourceCommand.Name {
				target.Commands[i].Flags = sourceCommand.Flags
			}
		}
	}
}

// addGlobalFlagsUsed adds the top-level flags to tasks where interpolation is used.
func addGlobalFlagsUsed(cmd *cli.Command, t *task.Task, cfg *config.Config) error {
	marshalled, err := yaml.Marshal(t)
	if err != nil {
		return err
	}

	for name, arg := range cfg.Args {
		arg.Name = name

		match, err := interp.Contains(marshalled, name)
		if err != nil {
			return err
		}

		if !match {
			continue
		}

		if err := addFlag(cmd, arg); err != nil {
			return errors.Wrapf(
				err,
				"could not add flag `%s` to command `%s`",
				arg.Name,
				t.Name,
			)
		}
	}

	return nil
}

func addFlag(command *cli.Command, arg *task.Arg) error {
	flag, err := task.CreateCLIFlag(arg)
	if err != nil {
		return err
	}
	command.Flags = append(command.Flags, flag)

	return nil
}
