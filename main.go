package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"gitlab.com/rliebz/tusk/appcli"
	"gitlab.com/rliebz/tusk/config"
	"gitlab.com/rliebz/tusk/ui"
)

func main() {
	globalFlagApp := appcli.NewSilentApp()

	var filename string
	globalFlagApp.Action = func(c *cli.Context) error {
		filename = c.String("file")
		ui.Verbose = c.Bool("verbose")
		return nil
	}

	// Only does partial parsing, so errors must be ignored
	globalFlagApp.Run(os.Args) // nolint: gas, errcheck

	cfgText, err := config.FindFile(filename)
	if err != nil {
		printErrorWithHelp(err)
		return
	}

	flagApp, err := appcli.NewFlagApp(cfgText)
	if err != nil {
		printErrorWithHelp(err)
		return
	}

	flags, ok := flagApp.Metadata["flagValues"].(map[string]string)
	if !ok {
		printErrorWithHelp(errors.New("could not read flags from metadata"))
		return
	}

	for flagName, value := range flags {
		pattern := config.InterpolationPattern(flagName)
		re, reErr := regexp.Compile(pattern)
		if reErr != nil {
			printErrorWithHelp(reErr)
			return
		}

		cfgText = re.ReplaceAll(cfgText, []byte(value))
	}

	app, err := appcli.NewExecutorApp(cfgText)
	if err != nil {
		printErrorWithHelp(err)
		return
	}

	appcli.CopyFlags(app, flagApp)

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
