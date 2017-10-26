package main

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/rliebz/tusk/appcli"
	"github.com/rliebz/tusk/ui"
)

var version = "dev"

// nolint: gocyclo
func main() {
	defer gracefulRecover()

	args := os.Args
	if args[len(args)-1] == appcli.CompletionFlag {
		ui.Verbosity = ui.VerbosityLevelSilent
	}

	meta, err := appcli.GetConfigMetadata(args)
	if err != nil {
		ui.Error(err)
		os.Exit(1)
	}

	if ui.Verbosity != ui.VerbosityLevelSilent {
		ui.Verbosity = meta.Verbosity
	}

	if err = os.Chdir(meta.Directory); err != nil {
		ui.Error(err)
		os.Exit(1)
	}

	if meta.PrintVersion && !meta.PrintHelp {
		ui.Print(version)
		os.Exit(0)
	}

	app, err := appcli.NewApp(meta)
	if err != nil {
		ui.Error(err)
		os.Exit(1)
	}

	if meta.PrintHelp {
		appcli.ShowAppHelp(app)
		os.Exit(0)
	}

	if err := app.Run(os.Args); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			ws := exitErr.Sys().(syscall.WaitStatus)
			os.Exit(ws.ExitStatus())
		} else {
			ui.Error(err)
			os.Exit(1)
		}
	}
}

func gracefulRecover() {
	if r := recover(); r != nil {
		ui.Error("recovered from panic: ", r)
		os.Exit(1)
	}
}
