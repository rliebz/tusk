package ui

import "fmt"

// PrintCommand prints the command to be executed.
func PrintCommand(command string) {
	message := fmt.Sprintf(
		"[%s] %s\n",
		blue(commandActionString),
		yellow(command),
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
