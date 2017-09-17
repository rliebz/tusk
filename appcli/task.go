package appcli

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/rliebz/tusk/config"
	"github.com/rliebz/tusk/config/task"
)

// addTasks adds a series of tasks to a cli.App using a command creator.
func addTasks(app *cli.App, cfg *config.Config, create commandCreator) error {
	for name, t := range cfg.Tasks {
		t.Name = name
		if err := addTask(app, cfg, t, create); err != nil {
			return errors.Wrapf(err, `could not add task "%s"`, t.Name)
		}
	}

	return nil
}

func addTask(app *cli.App, cfg *config.Config, t *task.Task, create commandCreator) error {
	command, err := create(app, t)
	if err != nil {
		return errors.Wrapf(err, `could not create command "%s"`, t.Name)
	}

	if err := config.AddSubTasks(cfg, t); err != nil {
		return errors.Wrap(err, "could not add sub-tasks")
	}

	if err := addAllFlagsUsed(cfg, command, t); err != nil {
		return errors.Wrap(err, "could not add flags")
	}

	app.Commands = append(app.Commands, *command)

	return nil
}
