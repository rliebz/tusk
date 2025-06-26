package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"syscall"

	"github.com/urfave/cli"

	"github.com/rliebz/tusk/appcli"
	"github.com/rliebz/tusk/runner"
	"github.com/rliebz/tusk/ui"
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
	logger := ui.New(ui.Config{
		Stdout: cfg.stdout,
		Stderr: cfg.stderr,
	})

	defer func() {
		if r := recover(); r != nil {
			logError(logger, cfg.args, fmt.Errorf("recovered from panic: %v", r))
			status = cmp.Or(status, 1)
		}
	}()

	meta, err := appcli.NewMetadata(logger, cfg.args)
	if err != nil && !appcli.IsCompleting(cfg.args) {
		logError(logger, cfg.args, err)
		return 1
	}

	status, err = runMeta(meta, cfg.args)
	if err != nil && appcli.IsCompleting(cfg.args) && meta.CfgPath != "" {
		// Try again without the config file to get global option completions
		status, err = runMeta(appcli.NewConfiglessMetadata(logger), cfg.args)
	}
	if err != nil {
		logError(logger, cfg.args, err)
		return cmp.Or(status, 1)
	}

	return status
}

func runMeta(meta *appcli.Metadata, args []string) (exitStatus int, err error) {
	switch {
	case appcli.IsCompleting(args):
	case meta.PrintHelp:
	case meta.PrintVersion:
		printVersion(meta)
		return 0, nil
	case meta.InstallCompletion != "":
		return 0, appcli.InstallCompletion(meta)
	case meta.UninstallCompletion != "":
		return 0, appcli.UninstallCompletion(meta)
	case meta.CleanCache:
		return 0, runner.CleanCache()
	case meta.CleanProjectCache:
		return 0, runner.CleanProjectCache(meta.CfgPath)
	}

	app, err := appcli.NewApp(args, meta)
	if err != nil {
		return 1, err
	}

	switch {
	case meta.PrintHelp:
		appcli.ShowAppHelp(meta.Logger, app)
		return 0, nil
	case meta.CleanTaskCache != "":
		if app.Command(meta.CleanTaskCache) == nil {
			return 0, fmt.Errorf("task %q is not defined", meta.CleanTaskCache)
		}
		return 0, runner.CleanTaskCache(meta.CfgPath, meta.CleanTaskCache)
	}

	return runApp(app, meta, args)
}

func printVersion(meta *appcli.Metadata) {
	if version == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			version = info.Main.Version
		}
	}

	meta.Logger.Println(version)
}

func runApp(app *cli.App, meta *appcli.Metadata, args []string) (int, error) {
	if err := app.Run(args); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if meta.Logger.Level() < ui.LevelVerbose {
				err = nil
			}
			ws := exitErr.Sys().(syscall.WaitStatus)
			return ws.ExitStatus(), err
		}

		return 1, err
	}

	return 0, nil
}

func logError(logger *ui.Logger, args []string, err error) {
	if !appcli.IsCompleting(args) {
		logger.Error(err)
	}
}
