package appcli

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"gitlab.com/rliebz/tusk/config"
	"gitlab.com/rliebz/tusk/ui"
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
		cli.BoolFlag{
			Name:  "version, V",
			Usage: "Print version and exit",
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
	cfg, err := config.Parse(cfgText)
	if err != nil {
		return nil, err
	}

	app := NewSilentApp()
	app.Metadata = make(map[string]interface{})
	app.Metadata["flagValues"] = make(map[string]string)

	if err = addTasks(app, cfg, createMetadataBuildCommand); err != nil {
		return nil, err
	}

	app.Action = func(c *cli.Context) error {
		ui.Verbose = c.Bool("verbose")
		if c.Bool("version") {
			ui.Print(c.App.Version)
			os.Exit(0)
		}
		return nil
	}

	return app, nil
}

// NewApp creates a cli.App that executes tasks.
func NewApp(cfgText []byte) (*cli.App, error) {
	flagApp, err := NewFlagApp(cfgText)
	if err != nil {
		return nil, err
	}

	if err = flagApp.Run(os.Args); err != nil {
		return nil, err
	}

	flags, ok := flagApp.Metadata["flagValues"].(map[string]string)
	if !ok {
		return nil, errors.New("could not read flags from metadata")
	}

	cfgText, err = config.Interpolate(cfgText, flags)
	if err != nil {
		return nil, err
	}

	cfg, err := config.Parse(cfgText)
	if err != nil {
		return nil, err
	}

	app := NewBaseApp()

	if err := addTasks(app, cfg, createExecuteCommand); err != nil {
		return nil, err
	}

	copyFlags(app, flagApp)

	return app, nil
}
