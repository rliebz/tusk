package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"gitlab.com/rliebz/tusk/config"
	"gitlab.com/rliebz/tusk/task"
	"gitlab.com/rliebz/tusk/ui"
)

func main() {
	app := createCLIApp()

	cfg, err := config.ReadTuskfile()
	if err != nil {
		printErrorWithHelp(err)
		return
	}

	if err := addTasks(app, cfg); err != nil {
		printErrorWithHelp(err)
		return
	}

	if err := app.Run(os.Args); err != nil {
		ui.PrintError(err)
	}
}

func createCLIApp() *cli.App {
	app := cli.NewApp()
	app.Usage = "a task runner built with simple configuration in mind"
	app.HideVersion = true
	app.HideHelp = true

	app.Flags = append(app.Flags,
		cli.HelpFlag,
		// The file flag will be read directly before calling `*cli.App#Run()`
		// It is only part of the cli.App for use with `tusk help`
		cli.StringFlag{
			Name:  "file, f",
			Usage: "Set `FILE` to use as the Tuskfile",
		},
	)

	return app
}

func addTasks(app *cli.App, cfg *config.Config) error {

	// Create commands
	for name, t := range cfg.Tasks {
		command, err := createCommand(name, t)
		if err != nil {
			return errors.Wrapf(err, "could not create command `%s`", name)
		}
		app.Commands = append(app.Commands, *command)
	}

	// Update pretasks
	for _, t := range cfg.Tasks {
		for _, name := range t.PreName {
			t.PreTasks = append(t.PreTasks, cfg.Tasks[name])
		}
	}

	return nil
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

func printErrorWithHelp(err error) {
	ui.PrintError(err)
	fmt.Println()
	showDefaultHelp()
}

func showDefaultHelp() {
	defaultApp := createCLIApp()
	context := cli.NewContext(defaultApp, nil, nil)
	if err := cli.ShowAppHelp(context); err != nil {
		ui.PrintError(err)
	}
}
