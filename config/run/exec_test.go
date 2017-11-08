package run

import (
	"bytes"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/rliebz/tusk/ui"
)

func TestExecCommand(t *testing.T) {
	command := "exit 0"

	stderrBuf := new(bytes.Buffer)
	ui.LoggerStderr.SetOutput(stderrBuf)
	ui.PrintCommand(command)
	stderrExpected := stderrBuf.String()

	stderrActualBuf := new(bytes.Buffer)
	ui.LoggerStderr.SetOutput(stderrActualBuf)
	if err := ExecCommand(command); err != nil {
		t.Fatalf(`execCommand("%s"): unexpected err: %s`, command, err)
	}
	stderrActual := stderrActualBuf.String()

	if stderrExpected != stderrActual {
		t.Errorf(
			"execCommand(\"%s\"):\nexpected stderr:\n`%s`\nactual:\n`%s`",
			command, stderrExpected, stderrActual,
		)
	}
}

func TestExecCommand_error(t *testing.T) {

	command := "exit 1"

	bufExpected := new(bytes.Buffer)
	errExpected := errors.New("exit status 1")
	ui.LoggerStderr.SetOutput(bufExpected)
	ui.PrintCommand(command)
	ui.PrintCommandError(errExpected)

	expected := bufExpected.String()

	bufActual := new(bytes.Buffer)
	ui.LoggerStderr.SetOutput(bufActual)
	if err := ExecCommand(command); err.Error() != errExpected.Error() {
		t.Fatalf(`execCommand("%s"): expected error "%s", actual "%s"`,
			command, errExpected, err,
		)
	}
	actual := bufActual.String()

	if expected != actual {
		t.Fatalf(
			"execCommand(\"%s\"):\nexpected output:\n`%s`\nactual output:\n`%s`",
			command, expected, actual,
		)
	}
}

func TestGetShell(t *testing.T) {
	originalShell := os.Getenv(shellEnvVar)
	defer func() {
		if err := os.Setenv(shellEnvVar, originalShell); err != nil {
			t.Errorf("Failed to reset SHELL environment variable: %v", err)
		}
	}()

	customShell := "/my/custom/sh"
	if err := os.Setenv(shellEnvVar, customShell); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}

	if actual := getShell(); actual != customShell {
		t.Errorf("getShell(): expected %v, actual %v", customShell, actual)
	}

	if err := os.Unsetenv(shellEnvVar); err != nil {
		t.Fatalf("Failed to unset environment variable: %v", err)
	}

	if actual := getShell(); actual != defaultShell {
		t.Errorf("getShell(): expected %v, actual %v", defaultShell, actual)
	}
}
