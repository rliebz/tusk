package ui

import (
	"fmt"
)

const (
	logFormat = "[%s] %s\n"

	debugString   = "Debug"
	infoString    = "Info"
	warningString = "Warning"
	errorString   = "Error"
)

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

func logInStyle(title string, f formatter, a ...interface{}) {
	message := fmt.Sprint(a...)
	printf(
		LoggerStderr,
		logFormat,
		f(title),
		message,
	)
}
