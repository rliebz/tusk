package appcli

import (
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/urfave/cli"

	"github.com/rliebz/tusk/runner"
)

// newBaseApp creates a basic cli.App with top-level flags.
func newBaseApp() *cli.App {
	app := cli.NewApp()
	app.Action = helpAction
	app.Usage = "the modern task runner"
	app.HideVersion = true
	app.HideHelp = true
	app.EnableBashCompletion = true
	app.UseShortOptionHandling = true
	app.ExitErrHandler = func(*cli.Context, error) {}

	app.Flags = append(app.Flags,
		// Options
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

		// Commands
		cli.BoolFlag{
			Name:  "h, help",
			Usage: "Show help and exit",
		},
		cli.BoolFlag{
			Name:  "V, version",
			Usage: "Print version and exit",
		},
		cli.StringFlag{
			Name:  "install-completion",
			Usage: "Install tab completion for a `shell` (one of: bash, fish, zsh)",
		},
		cli.StringFlag{
			Name:  "uninstall-completion",
			Usage: "Uninstall tab completion for a `shell` (one of: bash, fish, zsh)",
		},
		cli.BoolFlag{
			Name:  "clean-cache",
			Usage: "Delete all cached files",
		},
		cli.BoolFlag{
			Name:  "clean-project-cache",
			Usage: "Delete cached files related to the current config file",
		},
		cli.StringFlag{
			Name:  "clean-task-cache",
			Usage: "Delete cached files related to the given task",
		},
	)

	sort.Sort(cli.FlagsByName(app.Flags))
	return app
}

// newSilentApp creates a cli.App that will never print to stderr / stdout.
func newSilentApp() *cli.App {
	app := newBaseApp()
	app.Writer = io.Discard
	app.ErrWriter = io.Discard
	app.CommandNotFound = func(*cli.Context, string) {}
	return app
}

// newMetaApp creates a cli.App containing metadata, which can parse flags.
func newMetaApp(cfgText []byte) (*cli.App, error) {
	cfg, err := runner.Parse(cfgText)
	if err != nil {
		return nil, err
	}

	app := newSilentApp()
	app.Metadata = make(map[string]any)
	app.Metadata["tasks"] = make(map[string]*runner.Task)
	app.Metadata["argsPassed"] = []string{}
	app.Metadata["flagsPassed"] = make(map[string]string)

	if err := addTasks(app, nil, cfg, createMetadataBuildCommand); err != nil {
		return nil, err
	}

	return app, nil
}

func helpAction(c *cli.Context) error {
	if args := c.Args(); args.Present() {
		return fmt.Errorf("task %q is not defined", args.First())
	}

	cli.ShowAppHelp(c) //nolint:errcheck
	return nil
}

// NewApp creates a cli.App that executes tasks.
func NewApp(args []string, meta *runner.Metadata) (*cli.App, error) {
	metaApp, err := newMetaApp(meta.CfgText)
	if err != nil {
		return nil, err
	}

	if rerr := metaApp.Run(args); rerr != nil {
		return nil, rerr
	}

	var taskName string

	if command, ok := metaApp.Metadata["command"].(*cli.Command); ok {
		taskName = command.Name
	}

	argsPassed, flagsPassed, err := getPassedValues(metaApp)
	if err != nil {
		return nil, err
	}

	cfg, err := runner.ParseComplete(meta, taskName, argsPassed, flagsPassed)
	if err != nil {
		return nil, err
	}

	app := newBaseApp()
	if cfg.Name != "" {
		app.Name = cfg.Name
		app.HelpName = cfg.Name
	}
	if cfg.Usage != "" {
		app.Usage = cfg.Usage
	}
	if meta.Logger != nil {
		app.Writer = meta.Logger.Stdout
		app.ErrWriter = meta.Logger.Stderr
	}

	if err := addTasks(app, meta, cfg, createExecuteCommand); err != nil {
		return nil, err
	}

	copyFlags(app, metaApp)

	app.BashComplete = createDefaultComplete(meta.Logger.Stdout, app)
	for i := range app.Commands {
		cmd := &app.Commands[i]
		cmd.BashComplete = createCommandComplete(meta.Logger.Stdout, cmd, cfg)
	}

	return app, nil
}

// getPassedValues returns the args and flags passed by command line.
func getPassedValues(app *cli.App) (args []string, flags map[string]string, err error) {
	argsPassed, ok := app.Metadata["argsPassed"].([]string)
	if !ok {
		return nil, nil, errors.New("could not read args from metadata")
	}
	flagsPassed, ok := app.Metadata["flagsPassed"].(map[string]string)
	if !ok {
		return nil, nil, errors.New("could not read flags from metadata")
	}

	return argsPassed, flagsPassed, nil
}

// GetConfigMetadata returns a metadata object based on global options passed.
func GetConfigMetadata(args []string) (*runner.Metadata, error) {
	var err error
	app := newSilentApp()
	metadata := runner.NewMetadata()

	app.Action = func(c *cli.Context) error {
		// To prevent app from exiting, app.Action must return nil on error.
		// The enclosing function will still return the error.
		err = metadata.Set(c)
		return nil
	}

	if runErr := populateMetadata(app, args); runErr != nil {
		return nil, runErr
	}

	return metadata, err
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
