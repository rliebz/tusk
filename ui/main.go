package ui

import (
	"github.com/fatih/color"
)

// Level describes the verbosity of output.
type Level int

const (
	// LevelSilent does not print any output to stderr/stdout.
	LevelSilent Level = -8
	// LevelQuiet only prints command output and error messages.
	LevelQuiet Level = -4
	// LevelNormal is the normal level of verbosity.
	LevelNormal Level = 0
	// LevelVerbose prints all messages, include debug info.
	LevelVerbose Level = 4
)

const outputPrefix = " => "

func (v Level) String() string {
	switch v {
	case LevelSilent:
		return "Silent"
	case LevelQuiet:
		return "Quiet"
	case LevelNormal:
		return "Normal"
	case LevelVerbose:
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
		return name + ":"
	}

	return f(name)
}
