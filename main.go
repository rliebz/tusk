package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/rliebz/tusk/appcli"
	"github.com/rliebz/tusk/runner"
	"github.com/rliebz/tusk/ui"
	"github.com/urfave/cli"
)

var version = "dev"

func main() {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("recovered from panic: %v", r)
			unexpectedError(os.Args, err)
		}
	}()

	status := run(config{args: os.Args})
	os.Exit(status)
}

func unexpectedError(args []string, err error) {
	if !appcli.IsCompleting(args) {
		ui.New().Error(err)
	}
	os.Exit(1)
}

type config struct {
	args   []string
	stdout io.Writer
	stderr io.Writer
}

func run(cfg config) int {
	meta, err := appcli.GetConfigMetadata(cfg.args)
	if err != nil {
		unexpectedError(os.Args, err)
	}

	if cfg.stdout != nil {
		meta.Logger.Stdout = cfg.stdout
	}
	if cfg.stderr != nil {
		meta.Logger.Stderr = cfg.stderr
	}

	status, err := runMeta(meta, cfg.args)
	if err != nil {
		meta.Logger.Error(err)
		if status == 0 {
			return 1
		}
	}

	return status
}

func runMeta(meta *runner.Metadata, args []string) (exitStatus int, err error) {
	switch {
	case meta.InstallCompletion != "":
		return 0, appcli.InstallCompletion(meta)
	case meta.UninstallCompletion != "":
		return 0, appcli.UninstallCompletion(meta)
	case meta.PrintVersion && !meta.PrintHelp:
		meta.Logger.Println(version)
		return 0, nil
	}

	// TODO: Use runner.Context to avoid doing this
	if err = os.Chdir(meta.Directory); err != nil {
		return 1, err
	}

	app, err := appcli.NewApp(args, meta)
	if err != nil {
		return 1, err
	}

	if meta.PrintHelp {
		appcli.ShowAppHelp(meta.Logger, app)
		return 0, nil
	}

	return runApp(app, meta, args)
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
