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
		fmt.Sprintf("[%s] %s\n", commandActionString, "echo hello"),
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
