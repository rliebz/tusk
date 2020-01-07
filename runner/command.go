package runner

import (
	"os"
	"os/exec"

	"github.com/rliebz/tusk/marshal"
	"github.com/rliebz/tusk/ui"
)

const (
	shellEnvVar  = "SHELL"
	defaultShell = "sh"
)

// execCommand allows overwriting during tests.
var execCommand = exec.Command

// Command is a command passed to the shell.
type Command struct {
	Exec  string `yaml:"exec"`
	Print string `yaml:"print"`
	Dir   string `yaml:"dir"`
}

// UnmarshalYAML allows strings to be interpreted as Do actions.
func (c *Command) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var do string
	doCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&do) },
		Assign: func() {
			*c = Command{
				Exec:  do,
				Print: do,
			}
		},
	}

	type commandType Command // Use new type to avoid recursion
	var commandItem commandType
	commandCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&commandItem) },
		Assign: func() {
			*c = Command(commandItem)
			if c.Print == "" {
				c.Print = c.Exec
			}
		},
	}

	return marshal.UnmarshalOneOf(doCandidate, commandCandidate)
}

// execCommand executes a shell command.
func (c *Command) exec() error {
	shell := getShell()
	cmd := execCommand(shell, "-c", c.Exec)
	cmd.Dir = c.Dir
	cmd.Stdin = os.Stdin
	if ui.Verbosity > ui.VerbosityLevelSilent {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

// CommandList is a list of commands with custom yaml unamrshaling.
type CommandList []Command

// UnmarshalYAML allows single items to be used as lists.
func (cl *CommandList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var commandSlice []Command
	sliceCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&commandSlice) },
		Assign:    func() { *cl = commandSlice },
	}

	var commandItem Command
	itemCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&commandItem) },
		Assign:    func() { *cl = CommandList{commandItem} },
	}

	return marshal.UnmarshalOneOf(sliceCandidate, itemCandidate)
}

// getShell returns the value of the `SHELL` environment variable, or `sh`.
func getShell() string {
	if shell := os.Getenv(shellEnvVar); shell != "" {
		return shell
	}

	return defaultShell
}
