package ui

import (
	"log"
	"os"

	"github.com/fatih/color"
)

const (
	logLevelSilent  = iota
	logLevelQuiet   = iota
	logLevelNormal  = iota
	logLevelVerbose = iota
)

var (
	logLevel = logLevelNormal

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
	return logLevel == logLevelSilent
}

// IsQuiet returns true iff output is limited (quiet ot silent logging level)
func IsQuiet() bool {
	return IsSilent() || logLevel == logLevelQuiet
}

// IsVerbose returns true iff usng verbose logging level
func IsVerbose() bool {
	return logLevel == logLevelVerbose
}

// SetSilent sets the logging kevel to silent
func SetSilent() {
	logLevel = logLevelSilent
}

// SetQuiet sets the logging kevel to quiet
func SetQuiet() {
	logLevel = logLevelQuiet
}

// SetVerbose sets the logging level to verbose
func SetVerbose() {
	logLevel = logLevelVerbose
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
