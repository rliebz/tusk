package appcli

import (
	"errors"
	"io/ioutil"
	"os"
	"sort"

	"github.com/urfave/cli"

	"github.com/rliebz/tusk/runner"
)

// newBaseApp creates a basic cli.App with top-level flags.
func newBaseApp() *cli.App {
	app := cli.NewApp()
	app.Usage = "the modern task runner"
	app.HideVersion = true
	app.HideHelp = true
	app.EnableBashCompletion = true
	app.UseShortOptionHandling = true
	app.ExitErrHandler = func(*cli.Context, error) {}

	app.Flags = append(app.Flags,
		cli.BoolFlag{
			Name:  "h, help",
			Usage: "Show help and exit",
		},
		cli.StringFlag{
			Name:  "f, file",
			Usage: "Set `file` to use as the config file",
		},
		cli.StringFlag{
			Name:  "install-completion",
			Usage: "Install tab completion for a `shell`",
		},
		cli.StringFlag{
			Name:  "uninstall-completion",
			Usage: "Uninstall tab completion for a `shell`",
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

	sort.Sort(cli.FlagsByName(app.Flags))
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

// newMetaApp creates a cli.App containing metadata, which can parse flags.
func newMetaApp(cfgText []byte) (*cli.App, error) {
	cfg, err := runner.Parse(cfgText)
	if err != nil {
		return nil, err
	}

	app := newSilentApp()
	app.Metadata = make(map[string]interface{})
	app.Metadata["tasks"] = make(map[string]*runner.Task)
	app.Metadata["argsPassed"] = []string{}
	app.Metadata["flagsPassed"] = make(map[string]string)

	if err := addTasks(app, nil, cfg, createMetadataBuildCommand); err != nil {
		return nil, err
	}

	return app, nil
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
	command, ok := metaApp.Metadata["command"].(*cli.Command)
	if ok {
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

	if err := addTasks(app, meta, cfg, createExecuteCommand); err != nil {
		return nil, err
	}

	copyFlags(app, metaApp)

	app.BashComplete = createDefaultComplete(os.Stdout, app)
	for i := range app.Commands {
		cmd := &app.Commands[i]
		cmd.BashComplete = createCommandComplete(os.Stdout, cmd, cfg)
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
