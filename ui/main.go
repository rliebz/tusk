package ui

import (
	"fmt"

	"github.com/fatih/color"
)

// VerbosityLevel describes the verbosity of output.
type VerbosityLevel int

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

const outputPrefix = " => "

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
	bold   = newFormatter(color.Bold)
	blue   = newFormatter(color.FgBlue)
	cyan   = newFormatter(color.FgCyan)
	green  = newFormatter(color.FgGreen)
	red    = newFormatter(color.FgRed)
	yellow = newFormatter(color.FgYellow)
)

type formatter func(a ...any) string

func newFormatter(value ...color.Attribute) formatter {
	return func(a ...any) string {
		return color.New(value...).SprintFunc()(a...)
	}
}

func tag(name string, f formatter) string {
	if color.NoColor {
		return fmt.Sprintf("%s:", name)
	}

	return f(name)
}
