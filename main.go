package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"gitlab.com/rliebz/tusk/config"
	"gitlab.com/rliebz/tusk/task"
	"gitlab.com/rliebz/tusk/ui"
)

func main() {
	fileFlagApp := newSilentApp()

	var filename string
	fileFlagApp.Action = func(c *cli.Context) error {
		filename = c.String("file")
		return nil
	}

	// Only does partial parsing, so errors must be ignored
	fileFlagApp.Run(os.Args) // nolint: gas, errcheck

	cfg, err := config.ReadTuskfile(filename)
	if err != nil {
		printErrorWithHelp(err)
		return
	}

	app := newBaseApp()

	if err := addTasks(app, cfg); err != nil {
		printErrorWithHelp(err)
		return
	}

	if err := app.Run(os.Args); err != nil {
		ui.Error(err)
	}
}

func newBaseApp() *cli.App {
	app := cli.NewApp()
	app.Usage = "a task runner built with simple configuration in mind"
	app.HideVersion = true
	app.HideHelp = true

	app.Flags = append(app.Flags,
		cli.HelpFlag,
		cli.StringFlag{
			Name:  "file, f",
			Usage: "Set `FILE` to use as the Tuskfile",
		},
	)

	return app
}

// newSilentApp returns an app that will never print to stderr / stdout.
func newSilentApp() *cli.App {
	app := newBaseApp()
	app.Writer = ioutil.Discard
	app.ErrWriter = ioutil.Discard
	app.CommandNotFound = func(c *cli.Context, command string) {}
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
	ui.Error(err)
	fmt.Println()
	showDefaultHelp()
}

func showDefaultHelp() {
	defaultApp := newBaseApp()
	context := cli.NewContext(defaultApp, nil, nil)
	if err := cli.ShowAppHelp(context); err != nil {
		ui.Error(err)
	}
}
