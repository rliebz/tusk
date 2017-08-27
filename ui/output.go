package ui

import (
	"fmt"

	"github.com/fatih/color"
)

const (
	debugString   = "DEBUG"
	infoString    = "INFO"
	warningString = "WARNING"
	errorString   = "ERROR"
)

var (
	// HasPrinted indicates whether any output has been printed to the console.
	// This can be used to determine if a blank line should be printed before
	// new output.
	HasPrinted = false

	// Verbose enables verbose output.
	Verbose = false

	bold = color.New(color.Bold).SprintFunc()

	blue   = color.New(color.FgBlue).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
)

// Print prints a message
func Print(a ...interface{}) {
	message := fmt.Sprint(a...)
	println(Stdout, message)
}

// Debug prints info only in verbose mode.
func Debug(a ...interface{}) {
	if !Verbose {
		return
	}

	message := fmt.Sprint(a...)
	printf(
		Stderr,
		"[%s] %s\n",
		cyan(debugString),
		message,
	)
}

// Info prints application info.
func Info(a ...interface{}) {
	message := fmt.Sprint(a...)
	printf(
		Stderr,
		"[%s] %s\n",
		blue(infoString),
		message,
	)
}

// Warn prints an application warning.
func Warn(a ...interface{}) {
	message := fmt.Sprint(a...)
	printf(
		Stderr,
		"[%s] %s\n",
		yellow(warningString),
		message,
	)
}

// Error prints an application error.
func Error(a ...interface{}) {
	message := fmt.Sprint(a...)
	printf(
		Stderr,
		"[%s] %s\n",
		red(errorString),
		message,
	)
}
