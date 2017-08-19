package appcli

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"gitlab.com/rliebz/tusk/config"
	"gitlab.com/rliebz/tusk/task"
)

// addTasks adds a series of tasks to a cli.App using a command creator.
func addTasks(app *cli.App, cfg *config.Config, create commandCreator) error {
	for name, t := range cfg.Tasks {
		t.Name = name
		if err := addTask(app, cfg, t, create); err != nil {
			return errors.Wrapf(err, "could not add task `%s`", t.Name)
		}
	}

	return nil
}

func addTask(app *cli.App, cfg *config.Config, t *task.Task, create commandCreator) error {
	command, err := create(app, t)
	if err != nil {
		return errors.Wrapf(err, "could not create command `%s`", t.Name)
	}

	if err := addGlobalFlagsUsed(command, t, cfg); err != nil {
		return errors.Wrap(err, "could not add global args")
	}

	for _, pre := range t.Pre {
		// TODO: This requires tasks to be defined in order
		pt, ok := cfg.Tasks[pre.Name]
		if !ok {
			return fmt.Errorf("pre-task %s was referenced before definition", pre.Name)
		}

		if err := addGlobalFlagsUsed(command, pt, cfg); err != nil {
			return errors.Wrap(err, "could not add global args")
		}

		t.PreTasks = append(t.PreTasks, pt)
	}

	app.Commands = append(app.Commands, *command)

	return nil
}
