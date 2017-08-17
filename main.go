package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"gitlab.com/rliebz/tusk/appcli"
	"gitlab.com/rliebz/tusk/config"
	"gitlab.com/rliebz/tusk/ui"
)

func main() {
	meta := appcli.GetConfigMetadata(os.Args)

	ui.Verbose = meta.Verbose

	if meta.RunVersion {
		ui.Print("0.0.0")
		os.Exit(0)
	}

	cfgText, err := config.FindAndReadFile(meta.Filename)
	if err != nil {
		printErrorWithHelp(err)
		return
	}

	app, err := appcli.NewApp(cfgText)
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
	showDefaultHelp()
}

func showDefaultHelp() {
	defaultApp := appcli.NewBaseApp()
	context := cli.NewContext(defaultApp, nil, nil)
	if err := cli.ShowAppHelp(context); err != nil {
		ui.Error(err)
	}
}
