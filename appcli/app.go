package appcli

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/rliebz/tusk/config"
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
			Usage: "Show help",
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
			Name:  "v, verbose",
			Usage: "Print verbose output",
		},
		cli.BoolFlag{
			Name:  "V, version",
			Usage: "Print version and exit",
		},
	)

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
func NewApp(meta *config.Metadata) (*cli.App, error) {
	flagApp, err := newFlagApp(meta.CfgText)
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

	var taskName string
	command, ok := flagApp.Metadata["command"].(*cli.Command)
	if ok {
		taskName = command.Name
	}

	cfgText, flags, err := config.Interpolate(meta.CfgText, passed, taskName)
	if err != nil {
		return nil, err
	}

	cfg, err := config.Parse(cfgText)
	if err != nil {
		return nil, err
	}

	for _, t := range cfg.Tasks {
		t.Vars = flags
	}

	app := newBaseApp()

	if err := addTasks(app, cfg, createExecuteCommand); err != nil {
		return nil, err
	}

	copyFlags(app, flagApp)

	app.BashComplete = createBashComplete(app, meta)

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
		metadata.Quiet = c.Bool("quiet")
		metadata.Verbose = c.Bool("verbose")
		return err
	}

	if runErr := populateMetadata(app, metadata, args); runErr != nil {
		return nil, runErr
	}

	return metadata, err
}

func populateMetadata(app *cli.App, meta *config.Metadata, args []string) error {
	args, isCompleting := removeCompletionArg(args)
	meta.Completion.IsCompleting = isCompleting

	if err := app.Run(args); err != nil {
		if strings.HasPrefix(err.Error(), "flag needs an argument") {
			meta.Completion.IsFlagValue = true
			return app.Run(args[:len(args)-1])
		}
		return err
	}

	return nil
}
