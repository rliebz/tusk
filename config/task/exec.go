package task

import (
	"os"
	"os/exec"

	"github.com/rliebz/tusk/ui"
)

const shellEnvVar = "SHELL"
const defaultShell = "sh"

// execCommand executes a shell command.
func execCommand(command string) error {
	shell := getShell()
	cmd := exec.Command(shell, "-c", command) // nolint: gas
	cmd.Stdin = os.Stdin
	if ui.Verbosity > ui.VerbosityLevelSilent {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// getShell returns the value of the `SHELL` environment variable, or `sh`.
func getShell() string {
	if shell := os.Getenv(shellEnvVar); shell != "" {
		return shell
	}

	return defaultShell
}
