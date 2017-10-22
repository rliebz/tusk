package ui

import (
	"fmt"
)

const (
	debugString   = "DEBUG"
	infoString    = "INFO"
	warningString = "WARNING"
	errorString   = "ERROR"
)

// Print prints a message
func Print(a ...interface{}) {
	message := fmt.Sprint(a...)
	println(Stdout, message)
}

// Debug prints info only in verbose mode.
func Debug(a ...interface{}) {
	if !IsVerbose() {
		return
	}

	logInStyle(debugString, cyan, a...)
}

// Info prints application info.
func Info(a ...interface{}) {
	if IsQuiet() {
		return
	}

	logInStyle(infoString, blue, a...)
}

// Warn prints an application warning.
func Warn(a ...interface{}) {
	if IsQuiet() {
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
		Stderr,
		"[%s] %s\n",
		f(title),
		message,
	)
}
