package ui

import (
	"errors"
	"fmt"
	"testing"
)

var commandTests = []printTestCase{
	{
		`PrintCommand("echo hello")`,
		LoggerStderr,
		func() { PrintCommand("echo hello") },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		fmt.Sprintf("[%s] %s\n", runningString, "echo hello"),
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
			"[%s] %s\n%sset %s=%s\n%sset %s=%s\n%sunset %s\n",
			runningString,
			environmentMessage,
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
			"[%s] %s\n%s%s\n", skippedString, "echo hello", outputPrefix, "oops",
		),
	},
	{
		`PrintCommandError(errors.New("oops"))`,
		LoggerStderr,
		func() { PrintCommandError(errors.New("oops")) },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		fmt.Sprintf("%s%s\n", outputPrefix, "oops"),
	},
}

func TestCommandPrintFunctions(t *testing.T) {
	for _, tt := range commandTests {
		testPrint(t, tt)
	}
}
