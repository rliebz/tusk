package appcli

import (
	"io/ioutil"
	"os"

	"github.com/urfave/cli"

	"gitlab.com/rliebz/tusk/config"
)

// NewBaseApp creates a basic cli.App with top-level flags.
func NewBaseApp() *cli.App {
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
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Print verbose output",
		},
	)

	return app
}

// NewSilentApp creates a cli.App that will never print to stderr / stdout.
func NewSilentApp() *cli.App {
	app := NewBaseApp()
	app.Writer = ioutil.Discard
	app.ErrWriter = ioutil.Discard
	app.CommandNotFound = func(c *cli.Context, command string) {}
	return app
}

// NewFlagApp creates a cli.App that can parse flags.
func NewFlagApp(cfgText []byte) (*cli.App, error) {

	flagCfg, err := config.Parse(cfgText)
	if err != nil {
		return nil, err
	}

	flagApp := NewSilentApp()

	if err = addTasks(flagApp, flagCfg, createMetadataBuildCommand); err != nil {
		return nil, err
	}

	if err = flagApp.Run(os.Args); err != nil {
		return nil, err
	}

	return flagApp, nil
}

// NewExecutorApp creates a cli.App that executes tasks.
func NewExecutorApp(cfgText []byte) (*cli.App, error) {

	appCfg, err := config.Parse(cfgText)
	if err != nil {
		return nil, err
	}

	app := NewBaseApp()

	if err := addTasks(app, appCfg, createExecuteCommand); err != nil {
		return nil, err
	}

	return app, nil
}
