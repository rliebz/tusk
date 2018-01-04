package appcli

import (
	"fmt"
	"os"
	"strings"

	"github.com/rliebz/tusk/config"
	"github.com/urfave/cli"
)

// CompletionFlag is the flag passed when performing shell completions.
var CompletionFlag = "--" + cli.BashCompletionFlag.GetName()

// createDefaultComplete prints the completion metadata for the top-level app.
// The metadata includes the completion type followed by a list of options.
// The available completion types are "normal" and "file". Normal will return
// tasks and flags, while file allows completion engines to use system files.
func createDefaultComplete(app *cli.App) func(c *cli.Context) {
	return func(c *cli.Context) {
		if c.NArg() > 0 {
			return
		}

		_, isCompleting := getCompletingFlag(app.Flags, os.Args[len(os.Args)-2])
		if !isCompleting {
			fmt.Println("normal")
			for _, command := range app.Commands {
				printCommand(command)
			}
			for _, flag := range app.Flags {
				printFlag(c, flag)
			}
			return
		}

		// Default to file completion
		fmt.Println("file")
	}
}

// createCommandComplete prints the completion metadata for a cli command.
// The metadata includes the completion type followed by a list of options.
// The available completion types are "normal" and "file". Normal will return
// task-specific flags, while file allows completion engines to use system files.
func createCommandComplete(command *cli.Command, cfg *config.Config) func(c *cli.Context) {
	return func(c *cli.Context) {

		flagName, isCompleting := getCompletingFlag(command.Flags, os.Args[len(os.Args)-2])
		if !isCompleting {
			fmt.Println("task")
			for _, flag := range command.Flags {
				printFlag(c, flag)
			}
			return
		}

		t := cfg.Tasks[command.Name]
		opt, ok := t.Options[flagName]
		if !ok {
			opt, ok = cfg.Options[flagName]
			if !ok {
				return
			}
		}

		if len(opt.ValuesAllowed) > 0 {
			fmt.Println("value")
			for _, value := range opt.ValuesAllowed {
				fmt.Println(value)
			}
			return
		}

		// Default to file completion
		fmt.Println("file")
	}
}

func printCommand(command cli.Command) {
	if command.Hidden {
		return
	}

	if command.Usage == "" {
		fmt.Println(command.Name)
		return
	}

	fmt.Printf(
		"%s:%s\n",
		command.Name,
		strings.Replace(command.Usage, "\n", "", -1),
	)
}

func printFlag(c *cli.Context, flag cli.Flag) {
	values := strings.Split(flag.GetName(), ", ")
	for _, value := range values {
		if len(value) == 1 || c.IsSet(value) {
			continue
		}
		fmt.Printf(
			"--%s:%s\n",
			value,
			strings.Replace(getDescription(flag), "\n", "", -1),
		)
	}
}

func getDescription(flag cli.Flag) string {
	return strings.SplitN(flag.String(), "\t", 2)[1]
}

func removeCompletionArg(args []string) []string {
	var output []string
	for _, arg := range args {
		if arg != CompletionFlag {
			output = append(output, arg)
		}
	}

	return output
}

// getCompletingFlag tells if the trailing arg is an incomplete flag.
func getCompletingFlag(flags []cli.Flag, arg string) (string, bool) {

	if !strings.HasPrefix(arg, "-") {
		return "", false
	}

	name := strings.TrimLeft(arg, "-")

	for _, flag := range flags {
		if _, ok := flag.(cli.BoolFlag); ok {
			continue
		}

		if _, ok := flag.(cli.BoolTFlag); ok {
			continue
		}

		for _, candidate := range strings.Split(flag.GetName(), ", ") {
			if name == candidate {
				return name, true
			}
		}
	}

	return "", false
}

func isFlagArgumentError(err error) bool {
	return strings.HasPrefix(err.Error(), "flag needs an argument")
}
