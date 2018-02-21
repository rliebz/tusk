package ui

import (
	"fmt"
	"strings"
)

const (
	logFormat = "%s %s\n"

	debugString   = "Debug"
	infoString    = "Info"
	warningString = "Warning"
	errorString   = "Error"

	deprecatedString = "Deprecated"
)

// Store a list of sent deprecations to prevent duplicates
var deprecations []string

// Println prints a message to stdout.
func Println(a ...interface{}) {
	message := fmt.Sprint(a...)
	println(LoggerStdout, message)
}

// Debug prints info only in verbose mode.
func Debug(a ...interface{}) {
	if Verbosity < VerbosityLevelVerbose {
		return
	}

	logInStyle(debugString, cyan, a...)
}

// Info prints application info.
func Info(a ...interface{}) {
	if Verbosity <= VerbosityLevelQuiet {
		return
	}

	logInStyle(infoString, blue, a...)
}

// Warn prints an application warning.
func Warn(a ...interface{}) {
	if Verbosity <= VerbosityLevelQuiet {
		return
	}

	logInStyle(warningString, yellow, a...)
}

// Error prints an application error.
func Error(a ...interface{}) {
	logInStyle(errorString, red, a...)
}

// Deprecate prints a deprecation warning no more than once.
func Deprecate(a ...interface{}) {
	if Verbosity <= VerbosityLevelQuiet {
		return
	}

	if len(a) > 0 {
		message := fmt.Sprint(a[0])
		for _, d := range deprecations {

			if message == d {
				return
			}
		}
		deprecations = append(deprecations, message)
	}

	logInStyle(deprecatedString, yellow, a...)
	println(LoggerStderr)
}

func logInStyle(title string, f formatter, a ...interface{}) {
	messages := make([]string, 0, len(a))
	for _, message := range a {
		messages = append(messages, fmt.Sprint(message))
	}
	message := strings.Join(messages, fmt.Sprintf("\n%s", f(outputPrefix)))

	printf(
		LoggerStderr,
		logFormat,
		tag(title, f),
		message,
	)
}
