package ui

import "fmt"

const (
	commandActionString  = "Running"
	commandSkippedString = "Skipping"

	outputPrefix = "  =>"
)

// PrintCommand prints the command to be executed.
func PrintCommand(command string) {
	message := fmt.Sprintf(
		"[%s] %s\n",
		blue(commandActionString),
		bold(command),
	)
	stdout.Printf(message)
}

// PrintCommandSkipped prints the command skipped and the reason.
func PrintCommandSkipped(command string, reason string) {
	if !Verbose {
		return
	}

	message := fmt.Sprintf(
		"[%s] %s\n%s %s\n",
		yellow(commandSkippedString),
		bold(command),
		cyan(outputPrefix),
		reason,
	)
	stdout.Printf(message)
}

// PrintCommandOutput prints output from a running command.
func PrintCommandOutput(text string) {
	message := fmt.Sprintf(
		"%s %s\n",
		cyan(outputPrefix),
		text,
	)
	stdout.Printf(message)
}

// PrintCommandError prints an error from a running command.
func PrintCommandError(err error) {
	stderr.Printf(
		"%s [%s] %s\n",
		red(outputPrefix),
		red(errorString),
		err.Error(),
	)
}
