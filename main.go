package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/rliebz/tusk/appcli"
	"github.com/rliebz/tusk/ui"
)

var version = "dev"

func main() {
	status, err := run(os.Args)
	if err != nil {
		ui.Error(err)
		if status == 0 {
			status = 1
		}
	}
	os.Exit(status)
}

// nolint: gocyclo
func run(args []string) (exitStatus int, err error) {
	defer func() {
		if r := recover(); r != nil {
			exitStatus = 1
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	if args[len(args)-1] == appcli.CompletionFlag {
		ui.Verbosity = ui.VerbosityLevelSilent
	}

	meta, err := appcli.GetConfigMetadata(args)
	if err != nil {
		return 1, err
	}

	if ui.Verbosity != ui.VerbosityLevelSilent {
		ui.Verbosity = meta.Verbosity
	}

	switch {
	case meta.InstallCompletion != "":
		return 0, appcli.InstallCompletion(meta.InstallCompletion)
	case meta.UninstallCompletion != "":
		return 0, appcli.UninstallCompletion(meta.UninstallCompletion)
	case meta.PrintVersion && !meta.PrintHelp:
		ui.Println(version)
		return 0, nil
	}

	if err = os.Chdir(meta.Directory); err != nil {
		return 1, err
	}

	app, err := appcli.NewApp(args, meta)
	if err != nil {
		return 1, err
	}

	if meta.PrintHelp {
		appcli.ShowAppHelp(app)
		return 0, nil
	}

	if err := app.Run(args); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if ui.Verbosity < ui.VerbosityLevelVerbose {
				err = nil
			}
			ws := exitErr.Sys().(syscall.WaitStatus)
			return ws.ExitStatus(), err
		}

		return 1, err
	}

	return 0, nil
}
