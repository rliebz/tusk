package ui

import (
	"bytes"
	"os"
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
