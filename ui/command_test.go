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
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		"foo > bar $ echo hello\n",
	},
	{
		`PrintCommandWithParenthetical("echo hello", "paren", "foo", "bar")`,
		withStderr,
		func(l *Logger) { l.PrintCommandWithParenthetical("echo hello", "paren", "foo", "bar") },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
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
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
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
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		"",
	},
	{
		`PrintSkipped("echo hello", "oops")`,
		withStderr,
		func(l *Logger) { l.PrintSkipped("echo hello", "oops") },
		VerbosityLevelNormal,
		VerbosityLevelVerbose,
		fmt.Sprintf(
			"%s %s\n%s%s\n",
			tag(skippedString, yellow),
			"echo hello",
			outputPrefix,
			"oops",
		),
	},
	{
		`PrintTask("foo")`,
		withStderr,
		func(l *Logger) { l.PrintTask("foo") },
		VerbosityLevelNormal,
		VerbosityLevelVerbose,
		"Task Started: foo\n",
	},
	{
		`PrintTaskFinally("foo")`,
		withStderr,
		func(l *Logger) { l.PrintTaskFinally("foo") },
		VerbosityLevelNormal,
		VerbosityLevelVerbose,
		"Task Finally: foo\n",
	},
	{
		`PrintTaskCompleted("foo")`,
		withStderr,
		func(l *Logger) { l.PrintTaskCompleted("foo") },
		VerbosityLevelNormal,
		VerbosityLevelVerbose,
		"Task Completed: foo\n",
	},
	{
		`PrintCommandError(errors.New("oops"))`,
		withStderr,
		func(l *Logger) { l.PrintCommandError(errors.New("oops")) },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
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
