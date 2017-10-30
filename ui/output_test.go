package ui

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func resetUIState() {
	LoggerStdout.SetOutput(os.Stdout)
	LoggerStderr.SetOutput(os.Stderr)
	Verbosity = VerbosityLevelNormal
}

func TestPrint(t *testing.T) {
	defer resetUIState()

	stdoutBuf := new(bytes.Buffer)
	LoggerStdout.SetOutput(stdoutBuf)

	message := "Hello"
	expected := message + "\n"

	Println(message)
	actual := stdoutBuf.String()

	if expected != actual {
		t.Errorf(
			`Print("%s") with verbosity %v: expected "%s", actual: "%s"`,
			message, Verbosity, expected, actual,
		)
	}
}

func TestPrint_multi(t *testing.T) {
	defer resetUIState()

	stdoutBuf := new(bytes.Buffer)
	LoggerStdout.SetOutput(stdoutBuf)

	message1 := "Hello"
	message2 := "World!"
	expected := message1 + message2 + "\n"

	Println(message1, message2)
	actual := stdoutBuf.String()

	if expected != actual {
		t.Errorf(
			`Print("%s", "%s") with verbosity %v: expected "%s", actual: "%s"`,
			message1, message2, Verbosity, expected, actual,
		)
	}
}

func TestPrint_silent(t *testing.T) {
	defer resetUIState()

	stdoutBuf := new(bytes.Buffer)
	LoggerStdout.SetOutput(stdoutBuf)
	message := "Hello"

	Verbosity = VerbosityLevelSilent
	Println(message)
	actual := stdoutBuf.String()

	if "" != actual {
		t.Errorf(
			`Print("%s") with verbosity %v: expected no output, actual: %s`,
			message, Verbosity, actual,
		)
	}
}

type logLevelTestCase = struct {
	function       func(a ...interface{})
	levelNoOutput  VerbosityLevel
	levelOutput    VerbosityLevel
	logLevelString string
}

var logLevelTests = []logLevelTestCase{
	{Debug, VerbosityLevelNormal, VerbosityLevelVerbose, debugString},
	{Info, VerbosityLevelQuiet, VerbosityLevelNormal, infoString},
	{Warn, VerbosityLevelQuiet, VerbosityLevelNormal, warningString},
	{Error, VerbosityLevelSilent, VerbosityLevelQuiet, errorString},
}

func TestLogLevels(t *testing.T) {
	for _, tt := range logLevelTests {
		func(tt logLevelTestCase) {
			defer resetUIState()

			buf := new(bytes.Buffer)
			LoggerStderr.SetOutput(buf)
			message := "Hello"
			expected := fmt.Sprintf(
				logFormat,
				tt.logLevelString,
				message,
			)

			Verbosity = tt.levelOutput
			tt.function(message)
			actual := buf.String()

			if expected != actual {
				t.Errorf(
					`%s("%s") with verbosity %v: expected "%s", actual: "%s"`,
					strings.Title(tt.logLevelString),
					message,
					Verbosity,
					expected,
					actual,
				)
			}
		}(tt)
	}
}

func TestLogLevels_silent(t *testing.T) {
	for _, tt := range logLevelTests {
		func(tt logLevelTestCase) {
			defer resetUIState()

			buf := new(bytes.Buffer)
			LoggerStderr.SetOutput(buf)
			message := "Hello"

			Verbosity = tt.levelNoOutput
			tt.function(message)
			actual := buf.String()

			if "" != actual {
				t.Errorf(
					`%s("%s") with verbosity %v: expected no output, actual: %s`,
					strings.Title(tt.logLevelString),
					message,
					Verbosity,
					actual,
				)
			}
		}(tt)
	}
}
