package main

import (
	"os"

	"github.com/rliebz/tusk/appcli"
	"github.com/rliebz/tusk/ui"
)

func main() {
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
		// TODO: Determine when this should print
		ui.Error(err)
	}
}
