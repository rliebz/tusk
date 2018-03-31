package appcli

import (
	"fmt"
	"os"
	"strings"

	"github.com/rliebz/tusk/config"
	"github.com/rliebz/tusk/config/option"
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

		trailingArg := os.Args[len(os.Args)-2]
		isCompleting := isCompletingArg(app.Flags, trailingArg)
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
		t := cfg.Tasks[command.Name]
		trailingArg := os.Args[len(os.Args)-2]
		isCompleting := isCompletingArg(command.Flags, trailingArg)

		if !isCompleting {
			if len(c.Args())+1 <= len(t.Args) {
				fmt.Println("task-args")
				argName := t.OrderedArgNames[len(c.Args())]
				arg := t.Args[argName]
				for _, value := range arg.ValuesAllowed {
					fmt.Println(value)
				}
			} else {
				fmt.Println("task-no-args")
			}
			for _, flag := range command.Flags {
				printFlag(c, flag)
			}
			return
		}

		flagName := strings.TrimLeft(trailingArg, "-")

		options, err := config.FindAllOptions(t, cfg)
		if err != nil {
			return
		}

		opt, ok := getOptionFlag(flagName, options)
		if !ok {
			return
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

func getOptionFlag(flagName string, options []*option.Option) (*option.Option, bool) {
	for _, opt := range options {
		if flagName == opt.Name || flagName == opt.Short {
			return opt, true
		}
	}

	return nil, false
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

// isCompletingArg returns if the trailing arg is an incomplete flag.
func isCompletingArg(flags []cli.Flag, arg string) bool {

	if !strings.HasPrefix(arg, "-") {
		return false
	}

	name := strings.TrimLeft(arg, "-")
	short := len(arg) == len(name)+1

	for _, flag := range flags {
		if _, ok := flag.(cli.BoolFlag); ok {
			continue
		}

		if _, ok := flag.(cli.BoolTFlag); ok {
			continue
		}

		for _, candidate := range strings.Split(flag.GetName(), ", ") {
			if len(candidate) == 1 && !short {
				continue
			}
			if name == candidate {
				return true
			}
		}
	}

	return false
}

func isFlagArgumentError(err error) bool {
	return strings.HasPrefix(err.Error(), "flag needs an argument")
}
