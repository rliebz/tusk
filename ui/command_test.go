package ui

import (
	"errors"
	"fmt"
	"testing"
)

var commandTests = []printTestCase{
	{
		`PrintCommand("echo hello", "foo")`,
		LoggerStderr,
		func() { PrintCommand("echo hello", "foo") },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		"foo $ echo hello\n",
	},
	{
		`PrintCommandWithParenthetical("echo hello", "foo", "bar")`,
		LoggerStderr,
		func() { PrintCommandWithParenthetical("echo hello", "foo", "bar") },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		"foo (bar) $ echo hello\n",
	},
	{
		`PrintEnvironment()`,
		LoggerStderr,
		func() {
			a := "one"
			c := "three"

			PrintEnvironment(map[string]*string{
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
		LoggerStderr,
		func() { PrintEnvironment(nil) },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		"",
	},
	{
		`PrintSkipped("echo hello", "oops")`,
		LoggerStderr,
		func() { PrintSkipped("echo hello", "oops") },
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
		`PrintTask("foo", false)`,
		LoggerStderr,
		func() { PrintTask("foo", false) },
		VerbosityLevelNormal,
		VerbosityLevelVerbose,
		"Task Started: foo\n",
	},
	{
		`PrintTask("foo", true)`,
		LoggerStderr,
		func() { PrintTask("foo", true) },
		VerbosityLevelNormal,
		VerbosityLevelVerbose,
		"Sub-Task Started: foo\n",
	},
	{
		`PrintTaskFinally("foo", false)`,
		LoggerStderr,
		func() { PrintTaskFinally("foo", false) },
		VerbosityLevelNormal,
		VerbosityLevelVerbose,
		"Task Finally: foo\n",
	},
	{
		`PrintTaskFinally("foo", true)`,
		LoggerStderr,
		func() { PrintTaskFinally("foo", true) },
		VerbosityLevelNormal,
		VerbosityLevelVerbose,
		"Sub-Task Finally: foo\n",
	},
	{
		`PrintTaskCompleted("foo", false)`,
		LoggerStderr,
		func() { PrintTaskCompleted("foo", false) },
		VerbosityLevelNormal,
		VerbosityLevelVerbose,
		"Task Completed: foo\n",
	},
	{
		`PrintTaskCompleted(foo", true)`,
		LoggerStderr,
		func() { PrintTaskCompleted("foo", true) },
		VerbosityLevelNormal,
		VerbosityLevelVerbose,
		"Sub-Task Completed: foo\n",
	},
	{
		`PrintCommandError(errors.New("oops"))`,
		LoggerStderr,
		func() { PrintCommandError(errors.New("oops")) },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		fmt.Sprintf("%s\n", "oops"),
	},
}

func TestCommandPrintFunctions(t *testing.T) {
	for _, tt := range commandTests {
		testPrint(t, tt)
	}
}
