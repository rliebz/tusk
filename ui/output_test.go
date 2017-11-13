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
		fmt.Sprintf(logFormat, tag(debugString, cyan), "foo"),
	},
	{
		`Info("foo")`,
		LoggerStderr,
		func() { Info("foo") },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		fmt.Sprintf(logFormat, tag(infoString, blue), "foo"),
	},
	{
		`Warn("foo")`,
		LoggerStderr,
		func() { Warn("foo") },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		fmt.Sprintf(logFormat, tag(warningString, yellow), "foo"),
	},
	{
		`Error("foo")`,
		LoggerStderr,
		func() { Error("foo") },
		VerbosityLevelSilent,
		VerbosityLevelQuiet,
		fmt.Sprintf(logFormat, tag(errorString, red), "foo"),
	},
}

func TestPrintFunctions(t *testing.T) {
	for _, tt := range outputTests {
		testPrint(t, tt)
	}
}
