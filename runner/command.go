package runner

import (
	"os"
	"os/exec"

	"github.com/rliebz/tusk/marshal"
	"github.com/rliebz/tusk/ui"
)

var defaultInterpreter = []string{"sh", "-c"}

// execCommand allows overwriting during tests.
var execCommand = exec.Command

// Command is a command passed to the shell.
type Command struct {
	// Exec is the script to execute.
	Exec string `yaml:"exec"`

	// Print is the text that will be printed when the command is executed.
	Print string `yaml:"print"`

	// Quiet means that no text/hint will be printed before execution. Command
	// output is still printed, similar to '--quiet' flag.
	Quiet bool `yaml:"quiet,omitempty"`

	// Dir is the directory of the command.
	Dir string `yaml:"dir"`
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

// newCmd creates an exec.Cmd that uses the interpreter and the script passed.
func newCmd(ctx Context, script string) *exec.Cmd {
	interpreter := defaultInterpreter
	if len(ctx.Interpreter) > 0 {
		interpreter = ctx.Interpreter
	}

	path := interpreter[0]
	args := []string{script}
	if len(interpreter) > 1 {
		args = append(interpreter[1:], args...)
	}

	return execCommand(path, args...)
}

// execCommand executes a shell command.
func (c *Command) exec(ctx Context) error {
	cmd := newCmd(ctx, c.Exec)

	cmd.Dir = c.Dir
	cmd.Stdin = os.Stdin
	if ctx.Logger.Verbosity > ui.VerbosityLevelSilent {
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
