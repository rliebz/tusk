package ui

import (
	"errors"
	"fmt"
	"testing"
)

var commandTests = []printTestCase{
	{
		`PrintCommand("echo hello", "foo", "bar")`,
		withStderr,
		func(l *Logger) { l.PrintCommand("echo hello", "foo", "bar") },
		LevelQuiet,
		LevelNormal,
		"foo > bar $ echo hello\n",
	},
	{
		`PrintCommandWithParenthetical("echo hello", "paren", "foo", "bar")`,
		withStderr,
		func(l *Logger) { l.PrintCommandWithParenthetical("echo hello", "paren", "foo", "bar") },
		LevelQuiet,
		LevelNormal,
		"foo > bar (paren) $ echo hello\n",
	},
	{
		`PrintEnvironment()`,
		withStderr,
		func(l *Logger) {
			a := "one"
			c := "three"

			l.PrintEnvironment(map[string]*string{
				"A": &a,
				"B": nil,
				"C": &c,
			})
		},
		LevelQuiet,
		LevelNormal,
		fmt.Sprintf(
			"Setting Environment\n%sset %s=%s\n%sset %s=%s\n%sunset %s\n",
			outputPrefix, "A", "one",
			outputPrefix, "C", "three",
			outputPrefix, "B",
		),
	},
	{
		`PrintEnvironment(nil)`,
		withStderr,
		func(l *Logger) { l.PrintEnvironment(nil) },
		LevelQuiet,
		LevelNormal,
		"",
	},
	{
		`PrintCommandSkipped("echo hello", "oops")`,
		withStderr,
		func(l *Logger) { l.PrintCommandSkipped("echo hello", "oops") },
		LevelNormal,
		LevelVerbose,
		fmt.Sprintf(
			"%s %s\n%s%s\n",
			tag(skippedCommandString, yellow),
			"echo hello",
			outputPrefix,
			"oops",
		),
	},
	{
		`PrintTaskSkipped("echo hello", "oops")`,
		withStderr,
		func(l *Logger) { l.PrintTaskSkipped("my-task", "oops") },
		LevelNormal,
		LevelVerbose,
		fmt.Sprintf(
			"%s %s\n%s%s\n",
			tag(skippedTaskString, yellow),
			"my-task",
			outputPrefix,
			"oops",
		),
	},
	{
		`PrintTask("foo")`,
		withStderr,
		func(l *Logger) { l.PrintTask("foo") },
		LevelNormal,
		LevelVerbose,
		"Task Started: foo\n",
	},
	{
		`PrintTaskFinally("foo")`,
		withStderr,
		func(l *Logger) { l.PrintTaskFinally("foo") },
		LevelNormal,
		LevelVerbose,
		"Task Finally: foo\n",
	},
	{
		`PrintTaskCompleted("foo")`,
		withStderr,
		func(l *Logger) { l.PrintTaskCompleted("foo") },
		LevelNormal,
		LevelVerbose,
		"Task Completed: foo\n",
	},
	{
		`PrintCommandError(errors.New("oops"))`,
		withStderr,
		func(l *Logger) { l.PrintCommandError(errors.New("oops")) },
		LevelQuiet,
		LevelNormal,
		"oops\n",
	},
}

func TestCommandPrintFunctions(t *testing.T) {
	for _, tt := range commandTests {
		t.Run(tt.name, func(t *testing.T) {
			testPrint(t, tt)
		})
	}
}
