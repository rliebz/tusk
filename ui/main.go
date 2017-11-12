package ui

import (
	"log"
	"os"

	"github.com/fatih/color"
)

// VerbosityLevel describes the verbosity of output.
type VerbosityLevel uint8

const (
	// VerbosityLevelSilent does not print any output to stderr/stdout.
	VerbosityLevelSilent VerbosityLevel = iota
	// VerbosityLevelQuiet only prints command output and error messages.
	VerbosityLevelQuiet VerbosityLevel = iota
	// VerbosityLevelNormal is the normal level of verbosity.
	VerbosityLevelNormal VerbosityLevel = iota
	// VerbosityLevelVerbose prints all messages, include debug info.
	VerbosityLevelVerbose VerbosityLevel = iota
)

const outputPrefix = "=> "

func (v VerbosityLevel) String() string {
	switch v {
	case VerbosityLevelSilent:
		return "Silent"
	case VerbosityLevelQuiet:
		return "Quiet"
	case VerbosityLevelNormal:
		return "Normal"
	case VerbosityLevelVerbose:
		return "Verbose"
	default:
		return "Unknown"
	}
}

var (
	// Verbosity allows the verbosity of output to be set.
	Verbosity = VerbosityLevelNormal

	// LoggerStdout is a logger that prints to stdout.
	LoggerStdout = log.New(os.Stdout, "", 0)

	// LoggerStderr is a logger that prints to stderr.
	LoggerStderr = log.New(os.Stderr, "", 0)

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
}

func printf(l *log.Logger, format string, v ...interface{}) {
	if Verbosity == VerbosityLevelSilent {
		return
	}

	l.Printf(format, v...)
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
