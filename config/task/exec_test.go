package task

import (
	"os"
	"testing"

	"github.com/pkg/errors"
)

func TestExecCommand(t *testing.T) {
	command := "exit 0"

	if err := execCommand(command); err != nil {
		t.Fatalf(`execCommand("%s"): unexpected err: %s`, command, err)
	}
}

func TestExecCommand_error(t *testing.T) {
	command := "exit 1"
	errExpected := errors.New("exit status 1")
	if err := execCommand(command); err.Error() != errExpected.Error() {
		t.Fatalf(`execCommand("%s"): expected error "%s", actual "%s"`,
			command, errExpected, err,
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
