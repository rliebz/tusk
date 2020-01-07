package appcli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rliebz/tusk/config"
	"github.com/urfave/cli"
)

// CompletionFlag is the flag passed when performing shell completions.
var CompletionFlag = "--" + cli.BashCompletionFlag.GetName()

// context represents the subset of *cli.Context required for flag completion.
type context interface {
	// IsSet checks if a flag was already set, meaning we no longer need to
	// complete it.
	IsSet(string) bool
	// NArg is the number of non-flag arguments. This is used to determine if a
	// sub command is being called.
	NArg() int
}

// createDefaultComplete prints the completion metadata for the top-level app.
// The metadata includes the completion type followed by a list of options.
// The available completion types are "normal" and "file". Normal will return
// tasks and flags, while file allows completion engines to use system files.
func createDefaultComplete(w io.Writer, app *cli.App) func(c *cli.Context) {
	return func(c *cli.Context) {
		defaultComplete(w, c, app)
	}
}

func defaultComplete(w io.Writer, c context, app *cli.App) {
	// If there's an arg, but we're not using command-completion, it's a user
	// error. There's nothing to complete.
	if c.NArg() > 0 {
		return
	}

	trailingArg := os.Args[len(os.Args)-2]
	if isCompletingFlagArg(app.Flags, trailingArg) {
		fmt.Fprintln(w, "file")
		return
	}

	fmt.Fprintln(w, "normal")
	for i := range app.Commands {
		printCommand(w, &app.Commands[i])
	}
	for _, flag := range app.Flags {
		printFlag(w, c, flag)
	}
}

// createCommandComplete prints the completion metadata for a cli command.
// The metadata includes the completion type followed by a list of options.
// The available completion types are "normal" and "file". Normal will return
// task-specific flags, while file allows completion engines to use system files.
func createCommandComplete(w io.Writer, command *cli.Command, cfg *config.Config) func(c *cli.Context) {
	return func(c *cli.Context) {
		commandComplete(w, c, command, cfg)
	}
}

func commandComplete(w io.Writer, c context, command *cli.Command, cfg *config.Config) {
	t := cfg.Tasks[command.Name]
	trailingArg := os.Args[len(os.Args)-2]

	if isCompletingFlagArg(command.Flags, trailingArg) {
		printCompletingFlagArg(w, t, cfg, trailingArg)
		return
	}

	if c.NArg()+1 <= len(t.Args) {
		fmt.Fprintln(w, "task-args")
		arg := t.Args[c.NArg()]
		for _, value := range arg.ValuesAllowed {
			fmt.Fprintln(w, value)
		}
	} else {
		fmt.Fprintln(w, "task-no-args")
	}
	for _, flag := range command.Flags {
		printFlag(w, c, flag)
	}
}

func printCompletingFlagArg(w io.Writer, t *config.Task, cfg *config.Config, trailingArg string) {
	options, err := config.FindAllOptions(t, cfg)
	if err != nil {
		return
	}

	opt, ok := getOptionFlag(trailingArg, options)
	if !ok {
		return
	}

	if len(opt.ValuesAllowed) > 0 {
		fmt.Fprintln(w, "value")
		for _, value := range opt.ValuesAllowed {
			fmt.Fprintln(w, value)
		}
		return
	}

	// Default to file completion
	fmt.Fprintln(w, "file")
}

func getOptionFlag(flag string, options []*config.Option) (*config.Option, bool) {
	flagName := getFlagName(flag)
	for _, opt := range options {
		if flagName == opt.Name || flagName == opt.Short {
			return opt, true
		}
	}

	return nil, false
}

func printCommand(w io.Writer, command *cli.Command) {
	if command.Hidden {
		return
	}

	if command.Usage == "" {
		fmt.Fprintln(w, command.Name)
		return
	}

	fmt.Fprintf(
		w,
		"%s:%s\n",
		command.Name,
		strings.ReplaceAll(command.Usage, "\n", ""),
	)
}

func printFlag(w io.Writer, c context, flag cli.Flag) {
	values := strings.Split(flag.GetName(), ", ")
	for _, value := range values {
		if len(value) == 1 || c.IsSet(value) {
			continue
		}
		fmt.Fprintf(
			w,
			"--%s:%s\n",
			value,
			strings.ReplaceAll(getDescription(flag), "\n", ""),
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

// isCompletingFlagArg returns if the trailing arg is an incomplete flag.
func isCompletingFlagArg(flags []cli.Flag, arg string) bool {
	if !strings.HasPrefix(arg, "-") {
		return false
	}

	name := getFlagName(arg)
	short := !strings.HasPrefix(arg, "--")

	for _, flag := range flags {
		switch flag.(type) {
		case cli.BoolFlag, cli.BoolTFlag:
			continue
		}

		if flagMatchesName(flag, name, short) {
			return true
		}
	}

	return false
}

func flagMatchesName(flag cli.Flag, name string, short bool) bool {
	for _, candidate := range strings.Split(flag.GetName(), ", ") {
		if len(candidate) == 1 && !short {
			continue
		}

		if name == candidate {
			return true
		}
	}

	return false
}

func getFlagName(flag string) string {
	if strings.HasPrefix(flag, "--") {
		return flag[2:]
	}

	return flag[len(flag)-1:]
}

func isFlagArgumentError(err error) bool {
	return strings.HasPrefix(err.Error(), "flag needs an argument")
}
