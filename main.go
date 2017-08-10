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

type commandCreator func(app *cli.App, t *task.Task) (*cli.Command, error)

func main() {
	globalFlagApp := newSilentApp()

	var filename string
	globalFlagApp.Action = func(c *cli.Context) error {
		filename = c.String("file")
		ui.Verbose = c.Bool("verbose")
		return nil
	}

	// Only does partial parsing, so errors must be ignored
	globalFlagApp.Run(os.Args) // nolint: gas, errcheck

	cfgText, err := config.FindFile(filename)
	if err != nil {
		printErrorWithHelp(err)
		return
	}

	flagApp, err := newFlagApp(cfgText)
	if err != nil {
		printErrorWithHelp(err)
		return
	}

	flags, ok := flagApp.Metadata["flagValues"].(map[string]string)
	if !ok {
		printErrorWithHelp(errors.New("could not read flags from metadata"))
		return
	}

	for flagName, value := range flags {
		pattern := config.InterpolationPattern(flagName)
		re, reErr := regexp.Compile(pattern)
		if reErr != nil {
			printErrorWithHelp(reErr)
			return
		}

		cfgText = re.ReplaceAll(cfgText, []byte(value))
	}

	app := newBaseApp()

	appCfg, err := config.Parse(cfgText)
	if err != nil {
		printErrorWithHelp(err)
		return
	}

	if err := addTasks(app, appCfg, createExecuteCommand); err != nil {
		printErrorWithHelp(err)
		return
	}

	copyFlags(app, flagApp)

	if err := app.Run(os.Args); err != nil {
		ui.Error(err)
	}
}

func copyFlags(app *cli.App, flagApp *cli.App) {
	for i, command := range app.Commands {
		for _, flagCommand := range flagApp.Commands {
			if command.Name == flagCommand.Name {
				command.Flags = flagCommand.Flags
				app.Commands[i] = command
			}
		}
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
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Print verbose output",
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

func newFlagApp(cfgText []byte) (*cli.App, error) {

	flagCfg, err := config.Parse(cfgText)
	if err != nil {
		return nil, err
	}

	flagApp := newSilentApp()

	if err = addTasks(flagApp, flagCfg, createMetadataBuildCommand); err != nil {
		return nil, err
	}

	if err = flagApp.Run(os.Args); err != nil {
		return nil, err
	}

	return flagApp, nil
}

func addTasks(app *cli.App, cfg *config.Config, create commandCreator) error {
	for name, t := range cfg.Tasks {
		t.Name = name
		if err := addTask(app, cfg, t, create); err != nil {
			return errors.Wrapf(err, "could not add task `%s`", t.Name)
		}
	}

	return nil
}

func addTask(app *cli.App, cfg *config.Config, t *task.Task, create commandCreator) error {
	command, err := create(app, t)
	if err != nil {
		return errors.Wrapf(err, "could not create command `%s`", t.Name)
	}

	if err := addGlobalArgsUsed(command, t, cfg); err != nil {
		return errors.Wrap(err, "could not add global args")
	}

	for _, pre := range t.Pre {
		// TODO: This requires tasks to be defined in order
		pt := cfg.Tasks[pre.Name]
		if err := addGlobalArgsUsed(command, pt, cfg); err != nil {
			return errors.Wrap(err, "could not add global args")
		}
		// TODO: Tasks don't have `When` clauses, only scripts/pre-tasks do.
		// Tasks need `When` clauses
		// Also, the print should be moved to execution, not during build.
		if err := pre.When.Validate(); err != nil {
			// ui.PrintCommandSkipped("pre-task: "+pre.Name, err.Error())
		} else {
			t.PreTasks = append(t.PreTasks, pt)
		}
	}

	app.Commands = append(app.Commands, *command)

	return nil
}

func createExecuteCommand(app *cli.App, t *task.Task) (*cli.Command, error) {
	return createCommand(t, func(c *cli.Context) error {
		return t.Execute()
	})
}

func createMetadataBuildCommand(app *cli.App, t *task.Task) (*cli.Command, error) {
	if app.Metadata == nil {
		app.Metadata = make(map[string]interface{})
	}

	if app.Metadata["flagValues"] == nil {
		app.Metadata["flagValues"] = make(map[string]string)
	}

	flags, ok := app.Metadata["flagValues"].(map[string]string)
	if !ok {
		return nil, errors.New("could not read flags from metadata")
	}

	return createCommand(t, func(c *cli.Context) error {
		for _, flagName := range c.FlagNames() {
			flags[flagName] = c.String(flagName)
		}
		return nil
	})
}

// createCommand creates a cli.Command from a task.Task.
func createCommand(t *task.Task, actionFunc func(*cli.Context) error) (*cli.Command, error) {
	command := &cli.Command{
		Name:   t.Name,
		Usage:  t.Usage,
		Action: actionFunc,
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
