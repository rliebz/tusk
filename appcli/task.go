package appcli

import (
	"sort"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/rliebz/tusk/config"
)

// addTasks adds a series of tasks to a cli.App using a command creator.
func addTasks(app *cli.App, cfg *config.Config, create commandCreator) error {
	for _, t := range cfg.Tasks {
		if err := addTask(app, cfg, t, create); err != nil {
			return errors.Wrapf(err, `could not add task "%s"`, t.Name)
		}
	}

	sort.Sort(cli.CommandsByName(app.Commands))
	return nil
}

func addTask(app *cli.App, cfg *config.Config, t *config.Task, create commandCreator) error {
	if t.Private {
		return nil
	}

	command, err := create(app, t)
	if err != nil {
		return errors.Wrapf(err, `could not create command "%s"`, t.Name)
	}

	if err := addAllFlagsUsed(cfg, command, t); err != nil {
		return errors.Wrap(err, "could not add flags")
	}

	app.Commands = append(app.Commands, *command)

	return nil
}
