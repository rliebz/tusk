package ui

import (
	"fmt"
	"sort"
	"strings"
)

const (
	namespaceSeparator = " > "
	promptCharacter    = "$"

	completedString        = "Completed"
	environmentString      = "Setting Environment"
	finallyString          = "Finally"
	startedString          = "Started"
	setEnvironmentString   = "set"
	skippedString          = "Skipping"
	taskString             = "Task"
	unsetEnvironmentString = "unset"
)

// PrintCommand prints the command to be executed.
func (l Logger) PrintCommand(command string, namespaces ...string) {
	if l.Verbosity <= VerbosityLevelQuiet {
		return
	}

	for i, ns := range namespaces {
		namespaces[i] = green(ns)
	}

	s := strings.Join(namespaces, bold(blue(namespaceSeparator)))

	fmt.Fprintf(l.Stderr, "%s %s %s\n", s, bold(blue(promptCharacter)), bold(command))
}

// PrintCommandWithParenthetical prints a command with additional information.
func (l Logger) PrintCommandWithParenthetical(command, parenthetical string, namespaces ...string) {
	if l.Verbosity <= VerbosityLevelQuiet {
		return
	}

	for i, ns := range namespaces {
		namespaces[i] = green(ns)
	}

	s := strings.Join(namespaces, bold(blue(namespaceSeparator)))

	fmt.Fprintf(
		l.Stderr,
		"%s (%s) %s %s\n",
		s,
		yellow(parenthetical),
		bold(blue(promptCharacter)),
		bold(command),
	)
}

// PrintEnvironment prints when environment variables are set.
func (l Logger) PrintEnvironment(variables map[string]*string) {
	if l.Verbosity <= VerbosityLevelQuiet {
		return
	}

	if len(variables) == 0 {
		return
	}

	f := blue

	fmt.Fprintln(l.Stderr, f(environmentString))

	// Print in deterministic order
	keys := make([]string, 0, len(variables))
	for key := range variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := variables[key]
		if value == nil {
			continue
		}

		fmt.Fprintf(
			l.Stderr,
			"%s%s %s=%s\n",
			f(outputPrefix),
			setEnvironmentString,
			bold(key),
			*value,
		)
	}

	for _, key := range keys {
		value := variables[key]
		if value != nil {
			continue
		}

		fmt.Fprintf(
			l.Stderr,
			"%s%s %s\n",
			f(outputPrefix),
			unsetEnvironmentString,
			bold(key),
		)
	}
}

// PrintSkipped prints the command skipped and the reason.
//
// TODO: Split into PrintCommandSkipped and PrintTaskSkipped
func (l Logger) PrintSkipped(command, reason string) {
	if l.Verbosity < VerbosityLevelVerbose {
		return
	}

	f := cyan

	fmt.Fprintf(
		l.Stderr,
		logFormat,
		tag(skippedString, f),
		bold(command),
	)

	fmt.Fprintf(
		l.Stderr,
		"%s%s\n",
		f(outputPrefix),
		reason,
	)
}

// PrintTask prints when a task has begun.
func (l Logger) PrintTask(taskName string) {
	if l.Verbosity <= VerbosityLevelNormal {
		return
	}

	s := fmt.Sprintf("%s %s", taskString, startedString)

	fmt.Fprintf(
		l.Stderr,
		logFormat,
		tag(s, blue),
		bold(taskName),
	)
}

// PrintTaskFinally prints when a task's finally clause has begun.
func (l Logger) PrintTaskFinally(taskName string) {
	if l.Verbosity <= VerbosityLevelNormal {
		return
	}

	s := fmt.Sprintf("%s %s", taskString, finallyString)

	fmt.Fprintf(
		l.Stderr,
		logFormat,
		tag(s, blue),
		bold(taskName),
	)
}

// PrintTaskCompleted prints when a task has completed.
func (l Logger) PrintTaskCompleted(taskName string) {
	if l.Verbosity <= VerbosityLevelNormal {
		return
	}

	s := fmt.Sprintf("%s %s", taskString, completedString)

	fmt.Fprintf(
		l.Stderr,
		logFormat,
		tag(s, blue),
		bold(taskName),
	)
}

// PrintCommandError prints an error from a running command.
func (l Logger) PrintCommandError(err error) {
	if l.Verbosity <= VerbosityLevelQuiet {
		return
	}

	fmt.Fprintf(
		l.Stderr,
		"%s\n",
		red(err.Error()),
	)
}
