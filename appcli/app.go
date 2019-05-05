package appcli

import (
	"io/ioutil"
	"sort"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/rliebz/tusk/config"
	"github.com/rliebz/tusk/config/task"
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
			Name:   "install-completion",
			Usage:  "Install tab completion for a `shell`",
			Hidden: true,
		},
		cli.StringFlag{
			Name:   "uninstall-completion",
			Usage:  "Uninstall tab completion for a `shell`",
			Hidden: true,
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
	cfg, err := config.Parse(cfgText)
	if err != nil {
		return nil, err
	}

	app := newSilentApp()
	app.Metadata = make(map[string]interface{})
	app.Metadata["tasks"] = make(map[string]*task.Task)
	app.Metadata["argsPassed"] = []string{}
	app.Metadata["flagsPassed"] = make(map[string]string)

	if err := addTasks(app, cfg, createMetadataBuildCommand); err != nil {
		return nil, err
	}

	return app, nil
}

// NewApp creates a cli.App that executes tasks.
func NewApp(args []string, meta *config.Metadata) (*cli.App, error) {
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

	cfg, err := config.ParseComplete(meta.CfgText, taskName, argsPassed, flagsPassed)
	if err != nil {
		return nil, err
	}

	app := newBaseApp()
	if cfg.Name != nil {
		app.Name = *cfg.Name
		app.HelpName = *cfg.Name
	}
	if cfg.Usage != nil {
		app.Usage = *cfg.Usage
	}

	if err := addTasks(app, cfg, createExecuteCommand); err != nil {
		return nil, err
	}

	copyFlags(app, metaApp)

	app.BashComplete = createDefaultComplete(app)
	for i := range app.Commands {
		app.Commands[i].BashComplete = createCommandComplete(
			&app.Commands[i], cfg,
		)
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
func GetConfigMetadata(args []string) (*config.Metadata, error) {
	var err error
	app := newSilentApp()
	metadata := new(config.Metadata)

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
