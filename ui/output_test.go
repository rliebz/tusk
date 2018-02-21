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
		`Debug("foo", "bar", "baz")`,
		LoggerStderr,
		func() {
			Debug("foo", "bar", "baz")
		},
		VerbosityLevelNormal,
		VerbosityLevelVerbose,
		fmt.Sprintf(
			"%s %s\n%s%s\n%s%s\n",
			tag(debugString, cyan), "foo",
			cyan(outputPrefix), "bar",
			cyan(outputPrefix), "baz",
		),
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
	{
		`Deprecate("foo") once`,
		LoggerStderr,
		func() { Deprecate("foo") },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		fmt.Sprintf(logFormat, tag(deprecatedString, yellow), "foo\n"),
	},
	{
		`Deprecate("foo") twice`,
		LoggerStderr,
		func() {
			Deprecate("foo")
			Deprecate("foo")
		},
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		fmt.Sprintf(logFormat, tag(deprecatedString, yellow), "foo\n"),
	},
	{
		`Deprecate("foo", "bar")`,
		LoggerStderr,
		func() { Deprecate("foo", "bar") },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		fmt.Sprintf(
			"%s %s\n%s%s\n\n",
			tag(deprecatedString, yellow), "foo",
			yellow(outputPrefix), "bar",
		),
	},
}

func TestPrintFunctions(t *testing.T) {
	for _, tt := range outputTests {
		testPrint(t, tt)
	}
}
