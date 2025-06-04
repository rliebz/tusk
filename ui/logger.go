package ui

import (
	"fmt"
	"io"
	"os"
	"slices"
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

var (
	// Stdout is the default writer for stdout.
	Stdout io.Writer = os.Stdout
	// Stderr is the default writer for stderr.
	Stderr io.Writer = os.Stderr
)

// Logger writes CLI output at the appropriate level.
type Logger struct {
	Stdout, Stderr io.Writer
	Verbosity      VerbosityLevel

	deprecations []string
}

// New returns a new logger with the default settings.
func New() *Logger {
	return &Logger{
		Stdout:    Stdout,
		Stderr:    Stderr,
		Verbosity: VerbosityLevelNormal,
	}
}

// Noop returns a logger that does not print anything.
func Noop() *Logger {
	return &Logger{
		Stdout:    io.Discard,
		Stderr:    io.Discard,
		Verbosity: VerbosityLevelSilent,
	}
}

// Println prints a line directly.
func (l *Logger) Println(a ...any) {
	if l.Verbosity <= VerbosityLevelSilent {
		return
	}

	fmt.Fprintln(l.Stdout, a...)
}

// Debug prints debug information.
func (l *Logger) Debug(a ...any) {
	if l.Verbosity < VerbosityLevelVerbose {
		return
	}

	l.logInStyle(debugString, cyan, a...)
}

// Info prints normal application information.
func (l *Logger) Info(a ...any) {
	if l.Verbosity <= VerbosityLevelQuiet {
		return
	}

	l.logInStyle(infoString, blue, a...)
}

// Warn prints at the warning level.
func (l *Logger) Warn(a ...any) {
	if l.Verbosity <= VerbosityLevelQuiet {
		return
	}

	l.logInStyle(warningString, yellow, a...)
}

// Error prints application errors.
func (l *Logger) Error(a ...any) {
	if l.Verbosity <= VerbosityLevelSilent {
		return
	}

	l.logInStyle(errorString, red, a...)
}

// Deprecate prints deprecation warnings no more than once.
func (l *Logger) Deprecate(a ...any) {
	if l.Verbosity <= VerbosityLevelQuiet {
		return
	}

	if len(a) > 0 {
		message := fmt.Sprint(a[0])
		if slices.Contains(l.deprecations, message) {
			return
		}
		l.deprecations = append(l.deprecations, message)
	}

	l.logInStyle(deprecatedString, yellow, a...)
	fmt.Fprintln(l.Stderr)
}

func (l *Logger) logInStyle(title string, f formatter, a ...any) {
	messages := make([]string, 0, len(a))
	for _, message := range a {
		messages = append(messages, fmt.Sprint(message))
	}
	message := strings.Join(messages, "\n"+f(outputPrefix))

	fmt.Fprintf(l.Stderr, logFormat, tag(title, f), message)
}
