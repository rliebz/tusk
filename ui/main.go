package ui

import (
	"log"
	"os"

	"github.com/fatih/color"
)

var (
	// Silent prevents any output from printing.
	Silent = false

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

// IsSilent returns true iff using silent logging level
func IsSilent() bool {
	return Silent
}

// IsQuiet returns true iff output is limited (quiet ot silent logging level)
func IsQuiet() bool {
	return IsSilent() || Quiet
}

// IsVerbose returns true iff usng verbose logging level
func IsVerbose() bool {
	return !IsQuiet() && Verbose
}

func println(l *log.Logger, v ...interface{}) {
	if IsSilent() {
		return
	}

	l.Println(v...)
	HasPrinted = true
}

func printf(l *log.Logger, format string, v ...interface{}) {
	if IsSilent() {
		return
	}

	l.Printf(format, v...)
	HasPrinted = true
}

type formatter func(a ...interface{}) string

func conditionalColor(value ...color.Attribute) formatter {
	return func(a ...interface{}) string {
		if IsQuiet() {
			return color.New().SprintFunc()(a...)
		}

		return color.New(value...).SprintFunc()(a...)
	}
}
