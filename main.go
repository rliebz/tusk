package main

import (
	"os"

	"github.com/urfave/cli"

	"gitlab.com/rliebz/tusk/config"
	"gitlab.com/rliebz/tusk/task"
	"gitlab.com/rliebz/tusk/ui"
)

func createCLIApp(tuskfile *config.Config) (*cli.App, error) {
	app := cli.NewApp()
	app.Usage = "a task runner built with simple configuration in mind"
	app.HideVersion = true

	// This flag must be read directly before calling `*cli.App#Run()`
	// It is only part of the cli.App for use with `tusk help`
	app.Flags = append(app.Flags, cli.StringFlag{
		Name:  "file, f",
		Usage: "Set `FILE` to use as the Tuskfile",
	})

	taskMap := make(map[string]*task.Task)

	// Create commands
	for name, task := range tuskfile.Tasks {
		taskMap[name] = task
		command, err := createCommand(name, task)
		if err != nil {
			return nil, err
		}
		app.Commands = append(app.Commands, *command)
	}

	// Update pretasks
	for _, task := range tuskfile.Tasks {
		for _, name := range task.PreName {
			task.PreTasks = append(task.PreTasks, taskMap[name])
		}
	}

	return app, nil
}

func createCommand(name string, t *task.Task) (*cli.Command, error) {

	command := cli.Command{
		Name:  name,
		Usage: t.Usage,
		Action: func(c *cli.Context) error {
			return t.Execute()
		},
	}

	for name, arg := range t.Args {
		flag, err := task.CreateCLIFlag(name, arg)
		if err != nil {
			return nil, err
		}
		command.Flags = append(command.Flags, flag)
	}

	return &command, nil
}

func main() {
	// TODO: Show default help message for errors

	tuskfile, err := config.ReadTuskfile()
	if err != nil {
		ui.PrintError(err)
		return
	}

	app, err := createCLIApp(tuskfile)
	if err != nil {
		ui.PrintError(err)
		return
	}

	if err := app.Run(os.Args); err != nil {
		ui.PrintError(err)
	}
}
