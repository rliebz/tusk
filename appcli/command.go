package appcli

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"gitlab.com/rliebz/tusk/task"
)

type commandCreator func(app *cli.App, t *task.Task) (*cli.Command, error)

func createExecuteCommand(app *cli.App, t *task.Task) (*cli.Command, error) {
	return createCommand(t, func(c *cli.Context) error {
		return t.Execute()
	}), nil
}

func createMetadataBuildCommand(app *cli.App, t *task.Task) (*cli.Command, error) {
	flags, ok := app.Metadata["flagValues"].(map[string]string)
	if !ok {
		return nil, errors.New("could not read flags from metadata")
	}

	return createCommand(t, func(c *cli.Context) error {
		for _, flagName := range c.FlagNames() {
			if c.IsSet(flagName) {
				flags[flagName] = c.String(flagName)
			}
		}
		return nil
	}), nil
}

// createCommand creates a cli.Command from a task.Task.
func createCommand(t *task.Task, actionFunc func(*cli.Context) error) *cli.Command {
	return &cli.Command{
		Name:   t.Name,
		Usage:  t.Usage,
		Action: actionFunc,
	}
}
