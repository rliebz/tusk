package ui

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func resetUIState() {
	LoggerStdout.SetOutput(os.Stdout)
	LoggerStderr.SetOutput(os.Stderr)
	Verbosity = VerbosityLevelNormal
	deprecations = nil
}

type printTestCase struct {
	name            string
	logger          *log.Logger
	printFunc       func()
	levelNoOutput   VerbosityLevel
	levelWithOutput VerbosityLevel
	expected        string
}

func testPrint(t *testing.T, tt printTestCase) {
	t.Helper()
	defer resetUIState()

	buf := new(bytes.Buffer)
	tt.logger.SetOutput(buf)

	Verbosity = tt.levelNoOutput
	tt.printFunc()
	actual := buf.String()

	if actual != "" {
		t.Errorf(
			`%s with verbosity %v: expected no output, actual: "%s"`,
			tt.name,
			tt.levelNoOutput,
			actual,
		)
	}

	buf.Reset()
	Verbosity = tt.levelWithOutput
	tt.printFunc()
	actual = buf.String()

	if tt.expected != actual {
		t.Errorf(
			`%s with verbosity %v: expected "%s", actual: "%s"`,
			tt.name,
			tt.levelWithOutput,
			tt.expected,
			actual,
		)
	}
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
				`level.String(): expected "%s", actual "%s"`,
				tt.expected, actual,
			)
		}
	}
}
