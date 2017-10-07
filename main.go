package main

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/rliebz/tusk/appcli"
	"github.com/rliebz/tusk/ui"
)

var version = "dev"

func main() {
	defer func() {
		if r := recover(); r != nil {
			ui.Error("recovered from panic: ", r)
			os.Exit(1)
		}
	}()

	args := os.Args
	if args[len(args)-1] == appcli.CompletionFlag {
		ui.Silent = true
	}

	meta, err := appcli.GetConfigMetadata(args)
	if err != nil {
		ui.Error(err)
		appcli.ShowDefaultHelp()
		os.Exit(1)
	}

	ui.Quiet = meta.Quiet
	ui.Verbose = meta.Verbose
	if err = os.Chdir(meta.Directory); err != nil {
		ui.Error(err)
		os.Exit(1)
	}

	if meta.PrintHelp {
		appcli.ShowDefaultHelp()
		os.Exit(0)
	}

	if meta.PrintVersion {
		ui.Print(version)
		os.Exit(0)
	}

	app, err := appcli.NewApp(meta)
	if err != nil {
		ui.Error(err)
		appcli.ShowDefaultHelp()
		os.Exit(1)
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
