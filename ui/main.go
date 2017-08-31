package ui

import (
	"log"
	"os"

	"github.com/fatih/color"
)

var (
	// Quiet removes all additional formatting
	Quiet = false

	// Verbose enables verbose output.
	Verbose = false

	// HasPrinted indicates whether any output has been printed to the console.
	// This can be used to determine if a blank line should be printed before
	// new output.
	HasPrinted = false

	// Stdout is a logger that prints to stdout.
	Stdout = log.New(os.Stdout, "", 0)
	// Stderr is a logger that prints to stderr.
	Stderr = log.New(os.Stderr, "", 0)

	bold   = conditionalColor(color.Bold)
	blue   = conditionalColor(color.FgBlue)
	cyan   = conditionalColor(color.FgCyan)
	red    = conditionalColor(color.FgRed)
	yellow = conditionalColor(color.FgYellow)
)

func println(l *log.Logger, v ...interface{}) {
	l.Println(v...)
	HasPrinted = true
}

func printf(l *log.Logger, format string, v ...interface{}) {
	l.Printf(format, v...)
	HasPrinted = true
}

type formatter func(a ...interface{}) string

func conditionalColor(value ...color.Attribute) formatter {
	return func(a ...interface{}) string {
		if Quiet {
			return color.New().SprintFunc()(a...)
		}

		return color.New(value...).SprintFunc()(a...)
	}
}
