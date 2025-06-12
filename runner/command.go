package runner

import (
	"os"
	"os/exec"
	"path/filepath"

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
func (c *Command) UnmarshalYAML(unmarshal func(any) error) error {
	var str string
	strCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&str) },
		Assign: func() {
			*c = Command{
				Exec:  str,
				Print: str,
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

	return marshal.UnmarshalOneOf(strCandidate, commandCandidate)
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

	cmd := execCommand(path, args...)
	cmd.Dir = ctx.Dir()
	return cmd
}

// execCommand executes a shell command.
func (c *Command) exec(ctx Context) error {
	cmd := newCmd(ctx, c.Exec)

	cmd.Dir = filepath.Join(cmd.Dir, c.Dir)
	cmd.Stdin = os.Stdin
	if ctx.Logger.Level() > ui.VerbosityLevelSilent {
		cmd.Stdout = ctx.Logger.Stdout()
		cmd.Stderr = ctx.Logger.Stderr()
	}

	return cmd.Run()
}
