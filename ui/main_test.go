package ui

import (
	"bytes"
	"io"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func withStdout(l *Logger, out io.Writer) {
	l.Stdout = out
}

func withStderr(l *Logger, err io.Writer) {
	l.Stderr = err
}

type printTestCase struct {
	name            string
	setOutput       func(l *Logger, out io.Writer)
	printFunc       func(l *Logger)
	levelNoOutput   VerbosityLevel
	levelWithOutput VerbosityLevel
	expected        string
}

func testPrint(t *testing.T, tt printTestCase) {
	t.Helper()

	empty := new(bytes.Buffer)
	buf := new(bytes.Buffer)

	logger := New()
	logger.Stderr = empty
	logger.Stdout = empty
	tt.setOutput(logger, buf)

	logger.Verbosity = tt.levelNoOutput

	tt.printFunc(logger)
	actual := buf.String()

	if actual != "" {
		t.Errorf(
			"%s with verbosity %v: expected no output, actual: %q",
			tt.name,
			tt.levelNoOutput,
			actual,
		)
	}

	buf.Reset()

	logger.Verbosity = tt.levelWithOutput
	tt.printFunc(logger)
	actual = buf.String()

	if tt.expected != actual {
		t.Errorf(
			"%s with verbosity %v: expected %q, actual: %q",
			tt.name,
			tt.levelWithOutput,
			tt.expected,
			actual,
		)
	}

	assert.Check(t, cmp.Equal("", empty.String()), "fake")
}

var verbosityStringTests = []struct {
	level    VerbosityLevel
	expected string
}{
	{VerbosityLevelSilent, "Silent"},
	{VerbosityLevelQuiet, "Quiet"},
	{VerbosityLevelNormal, "Normal"},
	{VerbosityLevelVerbose, "Verbose"},
	{VerbosityLevel(99), "Unknown"},
}

func TestVerbosityLevel_String(t *testing.T) {
	for _, tt := range verbosityStringTests {
		actual := tt.level.String()
		if tt.expected != actual {
			t.Errorf(
				"level.String(): expected %q, actual %q",
				tt.expected, actual,
			)
		}
	}
}
