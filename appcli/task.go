package appcli

import (
	"fmt"
	"sort"

	"github.com/urfave/cli"

	"github.com/rliebz/tusk/runner"
)

// addTasks adds a series of tasks to a cli.App using a command creator.
func addTasks(
	app *cli.App,
	meta *runner.Metadata,
	cfg *runner.Config,
	create commandCreator,
) error {
	for _, t := range cfg.Tasks {
		if err := addTask(app, meta, cfg, t, create); err != nil {
			return fmt.Errorf("could not add task %q: %w", t.Name, err)
		}
	}

	sort.Sort(cli.CommandsByName(app.Commands))
	return nil
}

func addTask(
	app *cli.App,
	meta *runner.Metadata,
	cfg *runner.Config,
	t *runner.Task,
	create commandCreator,
) error {
	if t.Private {
		return nil
	}

	command, err := create(app, meta, t)
	if err != nil {
		return fmt.Errorf("could not create command %q: %w", t.Name, err)
	}

	options, err := runner.FindAllOptions(t, cfg)
	if err != nil {
		return fmt.Errorf("could not determine options for task %q: %w", t.Name, err)
	}

	if err := addAllFlagsUsed(command, t, options); err != nil {
		return fmt.Errorf("could not add flags for task %q: %w", t.Name, err)
	}

	command.CustomHelpTemplate = createCommandHelp(command, t, options)
	app.Commands = append(app.Commands, *command)

	return nil
}
