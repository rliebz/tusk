package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	yaml "gopkg.in/yaml.v2"

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

	cfgText, err := config.FindFile(filename)
	if err != nil {
		printErrorWithHelp(err)
		return
	}

	cfg, err := config.Parse(cfgText)
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
			Usage: "Set `FILE` to use as the config file",
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

// addTasks appends commands all tasks and pretasks to the app.
func addTasks(app *cli.App, cfg *config.Config) error {

	// Create commands
	for name, t := range cfg.Tasks {
		t.Name = name

		command, err := createCommand(t)
		if err != nil {
			return errors.Wrapf(err, "could not create command `%s`", t.Name)
		}

		if err := addGlobalArgsUsed(command, t, cfg); err != nil {
			return err
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

// createCommand creates a cli.Command from a task.Task.
func createCommand(t *task.Task) (*cli.Command, error) {
	command := &cli.Command{
		Name:  t.Name,
		Usage: t.Usage,
		Action: func(c *cli.Context) error {
			return t.Execute()
		},
	}

	for name, arg := range t.Args {
		arg.Name = name
		if err := addFlag(command, arg); err != nil {
			return nil, err
		}
	}

	return command, nil
}

// addGlobalArgsUsed adds the top-level args to tasks where interpolation is used.
func addGlobalArgsUsed(cmd *cli.Command, t *task.Task, cfg *config.Config) error {
	marshalled, err := yaml.Marshal(t)
	if err != nil {
		return err
	}

	for name, arg := range cfg.Args {
		arg.Name = name

		pattern := config.InterpolationPattern(arg.Name)
		match, err := regexp.Match(pattern, marshalled)
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

// TODO: Move to UI
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
