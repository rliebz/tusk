package appcli

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
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
		cli.BoolFlag{
			Name:  "version, V",
			Usage: "Print version and exit",
		},
	)

	return app
}

// newSilentApp creates a cli.App that will never print to stderr / stdout.
func newSilentApp() *cli.App {
	app := NewBaseApp()
	app.Writer = ioutil.Discard
	app.ErrWriter = ioutil.Discard
	app.CommandNotFound = func(c *cli.Context, command string) {}
	return app
}

// newFlagApp creates a cli.App that can parse flags.
func newFlagApp(cfgText []byte) (*cli.App, error) {
	cfg, err := config.Parse(cfgText)
	if err != nil {
		return nil, err
	}

	app := newSilentApp()
	app.Metadata = make(map[string]interface{})
	app.Metadata["flagsPassed"] = make(map[string]string)

	if err = addTasks(app, cfg, createMetadataBuildCommand); err != nil {
		return nil, err
	}

	return app, nil
}

// NewApp creates a cli.App that executes tasks.
func NewApp(cfgText []byte) (*cli.App, error) {
	flagApp, err := newFlagApp(cfgText)
	if err != nil {
		return nil, err
	}

	if err = flagApp.Run(os.Args); err != nil {
		return nil, err
	}

	passed, ok := flagApp.Metadata["flagsPassed"].(map[string]string)
	if !ok {
		return nil, errors.New("could not read flags from metadata")
	}

	cfgText, err = config.Interpolate(cfgText, passed)
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

// GetConfigMetadata returns a metadata object based on global flags.
func GetConfigMetadata(args []string) *config.Metadata {
	app := newSilentApp()

	metadata := new(config.Metadata)

	app.Action = func(c *cli.Context) error {
		metadata.Filename = c.String("file")
		metadata.Verbose = c.Bool("verbose")
		metadata.RunVersion = c.Bool("version")
		return nil
	}

	// Only does partial parsing, so errors must be ignored
	app.Run(args) // nolint: gas, errcheck

	return metadata
}
