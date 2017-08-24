package ui

const (
	commandActionString  = "Running"
	commandSkippedString = "Skipping"

	outputPrefix = "  =>"
)

// PrintCommand prints the command to be executed.
func PrintCommand(command string) {
	stderr.Printf(
		"[%s] %s\n",
		blue(commandActionString),
		bold(command),
	)
}

// PrintCommandSkipped prints the command skipped and the reason.
func PrintCommandSkipped(command string, reason string) {
	if !Verbose {
		return
	}

	stderr.Printf(
		"[%s] %s\n%s %s\n",
		yellow(commandSkippedString),
		bold(command),
		cyan(outputPrefix),
		reason,
	)
}

// PrintCommandOutput prints output from a running command.
func PrintCommandOutput(text string) {
	stderr.Printf(
		"%s %s\n",
		cyan(outputPrefix),
		text,
	)
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
