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
	cfgText, err := getConfigText(os.Args)
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

func getConfigText(args []string) ([]byte, error) {
	globalFlagApp := appcli.NewSilentApp()

	var filename string
	globalFlagApp.Action = func(c *cli.Context) error {
		filename = c.String("file")
		return nil
	}

	// Only does partial parsing, so errors must be ignored
	globalFlagApp.Run(args) // nolint: gas, errcheck

	return config.FindFile(filename)
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
