package main

import (
	"fmt"
	"os"

	"gitlab.com/rliebz/tusk/appcli"
	"gitlab.com/rliebz/tusk/ui"
)

func main() {
	meta, err := appcli.GetConfigMetadata(os.Args)
	if err != nil {
		printErrorWithHelp(err)
		return
	}

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
		printErrorWithHelp(err)
		return
	}

	if err := app.Run(os.Args); err != nil {
		ui.Error(err)
	}
}

// TODO: Move to UI
func printErrorWithHelp(err error) {
	ui.Error(err)
	fmt.Println()
	appcli.ShowDefaultHelp()
}
