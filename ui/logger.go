package ui

import (
	"cmp"
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

// Logger writes CLI output at the appropriate level.
type Logger struct {
	stdout, stderr io.Writer
	level          Level

	deprecations []string
}

// Config provides the configuration options for a [Logger].
type Config struct {
	Stdout    io.Writer
	Stderr    io.Writer
	Verbosity Level
}

// New returns a new logger with the default settings.
func New(cfg Config) *Logger {
	return &Logger{
		stdout: cfg.Stdout,
		stderr: cfg.Stderr,
		level:  cfg.Verbosity,
	}
}

// Noop returns a logger that does not print anything.
func Noop() *Logger {
	return &Logger{
		stdout: io.Discard,
		stderr: io.Discard,
		level:  LevelSilent,
	}
}

// Stdout returns the logger's standard output.
func (l *Logger) Stdout() io.Writer {
	return cmp.Or[io.Writer](l.stdout, os.Stdout)
}

// Stderr returns the logger's error output.
func (l *Logger) Stderr() io.Writer {
	return cmp.Or[io.Writer](l.stderr, os.Stderr)
}

// Level returns the logger's verbosity level.
func (l *Logger) Level() Level {
	return l.level
}

// SetLevel set the logger's verbosity level.
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// Println prints a line directly.
func (l *Logger) Println(a ...any) {
	if l.level <= LevelSilent {
		return
	}

	fmt.Fprintln(l.Stdout(), a...)
}

// Debug prints debug information.
func (l *Logger) Debug(a ...any) {
	if l.level < LevelVerbose {
		return
	}

	l.logInStyle(debugString, cyan, a...)
}

// Info prints normal application information.
func (l *Logger) Info(a ...any) {
	if l.level <= LevelQuiet {
		return
	}

	l.logInStyle(infoString, blue, a...)
}

// Warn prints at the warning level.
func (l *Logger) Warn(a ...any) {
	if l.level <= LevelQuiet {
		return
	}

	l.logInStyle(warningString, yellow, a...)
}

// Error prints application errors.
func (l *Logger) Error(a ...any) {
	if l.level <= LevelSilent {
		return
	}

	l.logInStyle(errorString, red, a...)
}

// Deprecate prints deprecation warnings no more than once.
func (l *Logger) Deprecate(a ...any) {
	if l.level <= LevelQuiet {
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
	fmt.Fprintln(l.Stderr())
}

func (l *Logger) logInStyle(title string, f formatter, a ...any) {
	messages := make([]string, 0, len(a))
	for _, message := range a {
		messages = append(messages, fmt.Sprint(message))
	}
	message := strings.Join(messages, "\n"+f(outputPrefix))

	fmt.Fprintf(l.Stderr(), logFormat, tag(title, f), message)
}
