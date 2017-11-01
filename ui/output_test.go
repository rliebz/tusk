package ui

import (
	"fmt"
	"testing"
)

var outputTests = []printTestCase{
	{
		`Println("foo", "bar")`,
		LoggerStdout,
		func() { Println("foo", "bar") },
		VerbosityLevelSilent,
		VerbosityLevelQuiet,
		"foobar\n",
	},
	{
		`Debug("foo")`,
		LoggerStderr,
		func() { Debug("foo") },
		VerbosityLevelNormal,
		VerbosityLevelVerbose,
		fmt.Sprintf(logFormat, debugString, "foo"),
	},
	{
		`Info("foo")`,
		LoggerStderr,
		func() { Info("foo") },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		fmt.Sprintf(logFormat, infoString, "foo"),
	},
	{
		`Warn("foo")`,
		LoggerStderr,
		func() { Warn("foo") },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		fmt.Sprintf(logFormat, warningString, "foo"),
	},
	{
		`Error("foo")`,
		LoggerStderr,
		func() { Error("foo") },
		VerbosityLevelSilent,
		VerbosityLevelQuiet,
		fmt.Sprintf(logFormat, errorString, "foo"),
	},
}

func TestPrintFunctions(t *testing.T) {
	for _, tt := range outputTests {
		testPrint(t, tt)
	}
}
