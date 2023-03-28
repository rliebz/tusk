package ui

import (
	"fmt"
	"testing"
)

var outputTests = []printTestCase{
	{
		`Println("foo", "bar")`,
		withStdout,
		func(l *Logger) { l.Println("foo", "bar") },
		VerbosityLevelSilent,
		VerbosityLevelQuiet,
		"foo bar\n",
	},
	{
		`Debug("foo")`,
		withStderr,
		func(l *Logger) { l.Debug("foo") },
		VerbosityLevelNormal,
		VerbosityLevelVerbose,
		fmt.Sprintf(logFormat, tag(debugString, cyan), "foo"),
	},
	{
		`Debug("foo", "bar", "baz")`,
		withStderr,
		func(l *Logger) { l.Debug("foo", "bar", "baz") },
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
		withStderr,
		func(l *Logger) { l.Info("foo") },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		fmt.Sprintf(logFormat, tag(infoString, blue), "foo"),
	},
	{
		`Warn("foo")`,
		withStderr,
		func(l *Logger) { l.Warn("foo") },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		fmt.Sprintf(logFormat, tag(warningString, yellow), "foo"),
	},
	{
		`Error("foo")`,
		withStderr,
		func(l *Logger) { l.Error("foo") },
		VerbosityLevelSilent,
		VerbosityLevelQuiet,
		fmt.Sprintf(logFormat, tag(errorString, red), "foo"),
	},
	{
		`Deprecate("foo") once`,
		withStderr,
		func(l *Logger) { l.Deprecate("foo") },
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		fmt.Sprintf(logFormat, tag(deprecatedString, yellow), "foo\n"),
	},
	{
		`Deprecate("foo") twice`,
		withStderr,
		func(l *Logger) {
			l.Deprecate("foo")
			l.Deprecate("foo")
		},
		VerbosityLevelQuiet,
		VerbosityLevelNormal,
		fmt.Sprintf(logFormat, tag(deprecatedString, yellow), "foo\n"),
	},
	{
		`Deprecate("foo", "bar")`,
		withStderr,
		func(l *Logger) { l.Deprecate("foo", "bar") },
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
		t.Run(tt.name, func(t *testing.T) {
			testPrint(t, tt)
		})
	}
}
