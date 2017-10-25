package ui

import (
	"log"
	"os"

	"github.com/fatih/color"
)

// VerbosityLevel describes amount of output
type VerbosityLevel uint8

const (
	// VerbosityLevelSilent means sending no output from tusk
	VerbosityLevelSilent VerbosityLevel = iota
	// VerbosityLevelQuiet means sending limited output from tusk
	VerbosityLevelQuiet VerbosityLevel = iota
	// VerbosityLevelNormal means normal output from tusk
	VerbosityLevelNormal VerbosityLevel = iota
	// VerbosityLevelVerbose means extra output from tusk
	VerbosityLevelVerbose VerbosityLevel = iota
)

var (
	// Verbosity is the amount to print for tusk output
	Verbosity = VerbosityLevelNormal

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
	if Verbosity == VerbosityLevelSilent {
		return
	}

	l.Println(v...)
	HasPrinted = true
}

func printf(l *log.Logger, format string, v ...interface{}) {
	if Verbosity == VerbosityLevelSilent {
		return
	}

	l.Printf(format, v...)
	HasPrinted = true
}

type formatter func(a ...interface{}) string

func conditionalColor(value ...color.Attribute) formatter {
	return func(a ...interface{}) string {
		if Verbosity <= VerbosityLevelQuiet {
			return color.New().SprintFunc()(a...)
		}

		return color.New(value...).SprintFunc()(a...)
	}
}
