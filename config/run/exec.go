package run

import (
	"os"
	"os/exec"

	"github.com/rliebz/tusk/ui"
)

// ExecCommand executes a shell command.
func ExecCommand(command string) error {
	ui.PrintCommand(command)

	cmd := exec.Command("sh", "-c", command) // nolint: gas
	cmd.Stdin = os.Stdin
	if ui.Verbosity > ui.VerbosityLevelSilent {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			ui.PrintCommandError(err)
		}
		return err
	}

	return nil
}
