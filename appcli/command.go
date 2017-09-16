package appcli

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/rliebz/tusk/config/configyaml/task"
)

type commandCreator func(app *cli.App, t *task.Task) (*cli.Command, error)

func createExecuteCommand(app *cli.App, t *task.Task) (*cli.Command, error) {
	return createCommand(t, func(c *cli.Context) error {
		return t.Execute()
	}), nil
}

func createMetadataBuildCommand(app *cli.App, t *task.Task) (*cli.Command, error) {
	passed, ok := app.Metadata["flagsPassed"].(map[string]string)
	if !ok {
		return nil, errors.New("could not read flags from metadata")
	}

	return createCommand(t, func(c *cli.Context) error {
		app.Metadata["command"] = &c.Command
		for _, flagName := range c.FlagNames() {
			if c.IsSet(flagName) {
				passed[flagName] = c.String(flagName)
			}
		}
		return nil
	}), nil
}

// createCommand creates a cli.Command from a task.Task.
func createCommand(t *task.Task, actionFunc func(*cli.Context) error) *cli.Command {
	return &cli.Command{
		Name:        t.Name,
		Usage:       strings.TrimSpace(t.Usage),
		Description: strings.TrimSpace(t.Description),
		Action:      actionFunc,
	}
}
