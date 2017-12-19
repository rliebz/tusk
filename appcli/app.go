package appcli

import (
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/rliebz/tusk/config"
	"github.com/rliebz/tusk/ui"
)

// newBaseApp creates a basic cli.App with top-level flags.
func newBaseApp() *cli.App {
	app := cli.NewApp()
	app.Usage = "a task runner built with simplicity in mind"
	app.HideVersion = true
	app.HideHelp = true
	app.EnableBashCompletion = true

	app.Flags = append(app.Flags,
		cli.BoolFlag{
			Name:  "h, help",
			Usage: "Show help and exit",
		},
		cli.StringFlag{
			Name:  "f, file",
			Usage: "Set `file` to use as the config file",
		},
		cli.BoolFlag{
			Name:  "q, quiet",
			Usage: "Only print command output and application errors",
		},
		cli.BoolFlag{
			Name:  "s, silent",
			Usage: "Print no output",
		},
		cli.BoolFlag{
			Name:  "v, verbose",
			Usage: "Print verbose output",
		},
		cli.BoolFlag{
			Name:  "V, version",
			Usage: "Print version and exit",
		},
	)

	sort.Sort(flagsByName(app.Flags))
	return app
}

// newSilentApp creates a cli.App that will never print to stderr / stdout.
func newSilentApp() *cli.App {
	app := newBaseApp()
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
func NewApp(args []string, meta *config.Metadata) (*cli.App, error) {
	flagApp, err := newFlagApp(meta.CfgText)
	if err != nil {
		return nil, err
	}

	if err = flagApp.Run(args); err != nil {
		return nil, err
	}

	passed, ok := flagApp.Metadata["flagsPassed"].(map[string]string)
	if !ok {
		return nil, errors.New("could not read flags from metadata")
	}

	var taskName string
	command, ok := flagApp.Metadata["command"].(*cli.Command)
	if ok {
		taskName = command.Name
	}

	cfg, err := config.ParseComplete(meta.CfgText, passed, taskName)
	if err != nil {
		return nil, err
	}

	app := newBaseApp()

	if err := addTasks(app, cfg, createExecuteCommand); err != nil {
		return nil, err
	}

	copyFlags(app, flagApp)

	app.BashComplete = createDefaultComplete(app)
	for i := range app.Commands {
		app.Commands[i].BashComplete = createCommandComplete(
			&app.Commands[i], cfg,
		)
	}

	return app, nil
}

// GetConfigMetadata returns a metadata object based on global options passed.
func GetConfigMetadata(args []string) (*config.Metadata, error) {

	var err error
	app := newSilentApp()
	metadata := new(config.Metadata)

	// To prevent app from exiting, the app.Action must return nil on error.
	// The enclosing function will still return the error.
	app.Action = func(c *cli.Context) error {
		fullPath := c.String("file")
		if fullPath != "" {
			metadata.CfgText, err = ioutil.ReadFile(fullPath)
			if err != nil {
				return nil
			}
		} else {
			var found bool
			fullPath, found, err = config.SearchForFile()
			if err != nil {
				return nil
			}

			if found {
				metadata.CfgText, err = ioutil.ReadFile(fullPath)
				if err != nil {
					return nil
				}
			}
		}

		metadata.Directory = filepath.Dir(fullPath)
		metadata.PrintHelp = c.Bool("help")
		metadata.PrintVersion = c.Bool("version")
		setMetadataVerbosity(metadata, c)

		return err
	}

	if runErr := populateMetadata(app, args); runErr != nil {
		return nil, runErr
	}

	return metadata, err
}

func setMetadataVerbosity(metadata *config.Metadata, c *cli.Context) {
	if c.Bool("silent") {
		metadata.Verbosity = ui.VerbosityLevelSilent
	} else if c.Bool("quiet") {
		metadata.Verbosity = ui.VerbosityLevelQuiet
	} else if c.Bool("verbose") {
		metadata.Verbosity = ui.VerbosityLevelVerbose
	} else {
		metadata.Verbosity = ui.VerbosityLevelNormal
	}
}

// populateMetadata runs the app to populate the metadata struct.
func populateMetadata(app *cli.App, args []string) error {
	args = removeCompletionArg(args)

	if err := app.Run(args); err != nil {
		// Ignore flags without arguments during metadata creation
		if isFlagArgumentError(err) {
			return app.Run(args[:len(args)-1])
		}

		return err
	}

	return nil
}
