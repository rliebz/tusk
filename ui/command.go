package ui

import (
	"fmt"
	"sort"
)

const (
	promptCharacter = "$"

	completedString        = "Completed"
	environmentString      = "Setting Environment"
	finallyString          = "Finally"
	runningString          = "Running"
	setEnvironmentString   = "set"
	skippedString          = "Skipping"
	subTaskString          = "Sub-Task"
	taskString             = "Task"
	unsetEnvironmentString = "unset"
)

// PrintCommand prints the command to be executed.
func PrintCommand(command, context string) {
	if Verbosity <= VerbosityLevelQuiet {
		return
	}

	printf(
		LoggerStderr,
		"%s %s %s",
		green(context),
		green(promptCharacter),
		bold(command),
	)
}

// PrintCommandWithParenthetical prints a command with additional information.
func PrintCommandWithParenthetical(command, context, parenthetical string) {
	if Verbosity <= VerbosityLevelQuiet {
		return
	}

	printf(
		LoggerStderr,
		"%s (%s) %s %s",
		green(context),
		yellow(parenthetical),
		green(promptCharacter),
		bold(command),
	)
}

// PrintEnvironment prints when environment variables are set.
func PrintEnvironment(variables map[string]*string) {
	if Verbosity <= VerbosityLevelQuiet {
		return
	}

	if len(variables) == 0 {
		return
	}

	f := blue

	println(
		LoggerStderr,
		f(environmentString),
	)

	// Print in deterministic order
	var keys []string
	for key := range variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := variables[key]
		if value == nil {
			continue
		}

		printf(
			LoggerStderr,
			"%s%s %s=%s",
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

		printf(
			LoggerStderr,
			"%s%s %s",
			f(outputPrefix),
			unsetEnvironmentString,
			bold(key),
		)
	}
}

// PrintSkipped prints the command skipped and the reason.
func PrintSkipped(command string, reason string) {
	if Verbosity < VerbosityLevelVerbose {
		return
	}

	f := cyan

	printf(
		LoggerStderr,
		logFormat,
		tag(skippedString, f),
		bold(command),
	)

	printf(
		LoggerStderr,
		"%s%s\n",
		f(outputPrefix),
		reason,
	)
}

// PrintTask prints when a task has begun.
func PrintTask(taskName string, asSubTask bool) {
	if Verbosity <= VerbosityLevelNormal {
		return
	}

	s := fmt.Sprintf("%s %s", getTaskString(asSubTask), runningString)

	printf(
		LoggerStderr,
		logFormat,
		tag(s, blue),
		bold(taskName),
	)
}

// PrintTaskFinally prints when a task's finally clause has begun.
func PrintTaskFinally(taskName string, asSubTask bool) {
	if Verbosity <= VerbosityLevelNormal {
		return
	}

	s := fmt.Sprintf("%s %s", getTaskString(asSubTask), finallyString)

	printf(
		LoggerStderr,
		logFormat,
		tag(s, blue),
		bold(taskName),
	)
}

// PrintTaskCompleted prints when a task has completed.
func PrintTaskCompleted(taskName string, asSubTask bool) {
	if Verbosity <= VerbosityLevelNormal {
		return
	}

	s := fmt.Sprintf("%s %s", getTaskString(asSubTask), completedString)

	printf(
		LoggerStderr,
		logFormat,
		tag(s, blue),
		bold(taskName),
	)
}

func getTaskString(asSubTask bool) string {
	if asSubTask {
		return subTaskString
	}

	return taskString
}

// PrintCommandError prints an error from a running command.
func PrintCommandError(err error) {
	if Verbosity <= VerbosityLevelQuiet {
		return
	}

	printf(
		LoggerStderr,
		"%s\n",
		red(err.Error()),
	)
}
