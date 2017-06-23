package ui

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
)

const (
	commandActionString = "Running"
	errorString         = "ERROR"
	outputPrefix        = "  =>"
)

var (
	stdout = log.New(os.Stdout, "", 0)
	stderr = log.New(os.Stderr, "", 0)

	blue   = color.New(color.FgBlue).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
)

// PrintError prints an application error.
func PrintError(err error) {
	stderr.Printf(
		"[%s] %s\n",
		red(errorString),
		err.Error(),
	)
}

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
	message := fmt.Sprintf("%s %s\n", cyan(outputPrefix), text)
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
