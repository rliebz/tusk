package main

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/rliebz/tusk/appcli"
	"github.com/rliebz/tusk/ui"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			ui.Error("recovered from panic: ", r)
			os.Exit(1)
		}
	}()

	meta, err := appcli.GetConfigMetadata(os.Args)
	if err != nil {
		ui.Error(err)
		appcli.ShowDefaultHelp()
		return
	}

	ui.Quiet = meta.Quiet
	ui.Verbose = meta.Verbose
	if err = os.Chdir(meta.Directory); err != nil {
		ui.Error(err)
		return
	}

	if meta.RunVersion {
		ui.Print("0.0.0")
		os.Exit(0)
	}

	app, err := appcli.NewApp(meta.CfgText)
	if err != nil {
		ui.Error(err)
		appcli.ShowDefaultHelp()
		return
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
