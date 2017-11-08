package run

import (
	"os"
	"os/exec"

	"github.com/rliebz/tusk/ui"
)

const shellEnvVar = "SHELL"
const defaultShell = "/bin/sh"

// ExecCommand executes a shell command.
func ExecCommand(command string) error {
	ui.PrintCommand(command)

	shell := getShell()
	cmd := exec.Command(shell, "-c", command) // nolint: gas
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

// getShell returns the value of the `SHELL` environment variable, or `/bin/sh`.
func getShell() string {
	if shell := os.Getenv(shellEnvVar); shell != "" {
		return shell
	}

	return defaultShell
}
