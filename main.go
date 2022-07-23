package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"syscall"

	"github.com/rliebz/tusk/appcli"
	"github.com/rliebz/tusk/runner"
	"github.com/rliebz/tusk/ui"
	"github.com/urfave/cli"
)

var version string

func main() {
	status := run(config{
		args:   os.Args,
		stdout: os.Stdout,
		stderr: os.Stderr,
	})
	os.Exit(status)
}

type config struct {
	args   []string
	stdout io.Writer
	stderr io.Writer
}

func run(cfg config) (status int) {
	defer func() {
		if r := recover(); r != nil {
			if !appcli.IsCompleting(cfg.args) {
				err := fmt.Errorf("recovered from panic: %v", r)
				ui.New().Error(err)
			}

			if status == 0 {
				status = 1
			}
		}
	}()

	meta, err := appcli.GetConfigMetadata(cfg.args)
	if err != nil {
		if !appcli.IsCompleting(cfg.args) {
			ui.New().Error(err)
		}
		return 1
	}

	meta.Logger.Stdout = cfg.stdout
	meta.Logger.Stderr = cfg.stderr

	status, err = runMeta(meta, cfg.args)
	if err != nil {
		meta.Logger.Error(err)
		if status == 0 {
			return 1
		}
	}

	return status
}

func runMeta(meta *runner.Metadata, args []string) (exitStatus int, err error) {
	printHelp := false

	switch {
	case appcli.IsCompleting(args):
	case meta.PrintHelp:
		printHelp = true
	case meta.PrintVersion:
		printVersion(meta)
		return 0, nil
	case meta.InstallCompletion != "":
		return 0, appcli.InstallCompletion(meta)
	case meta.UninstallCompletion != "":
		return 0, appcli.UninstallCompletion(meta)
	}

	// TODO: Use runner.Context to avoid doing this
	if err = os.Chdir(meta.Directory); err != nil {
		return 1, err
	}

	app, err := appcli.NewApp(args, meta)
	if err != nil {
		return 1, err
	}

	if printHelp {
		appcli.ShowAppHelp(meta.Logger, app)
		return 0, nil
	}

	return runApp(app, meta, args)
}

func printVersion(meta *runner.Metadata) {
	if version == "" {
		version = "(devel)"
		if info, ok := debug.ReadBuildInfo(); ok {
			version = info.Main.Version
		}
	}

	meta.Logger.Println(version)
}

func runApp(app *cli.App, meta *runner.Metadata, args []string) (int, error) {
	if err := app.Run(args); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if meta.Logger.Verbosity < ui.VerbosityLevelVerbose {
				err = nil
			}
			ws := exitErr.Sys().(syscall.WaitStatus)
			return ws.ExitStatus(), err
		}

		return 1, err
	}

	return 0, nil
}
