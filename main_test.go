package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/rliebz/tusk/ui"
)

func TestRun_PrintVersion(t *testing.T) {
	stdout, _, cleanup := withCapturedOutput()
	defer cleanup()

	args := []string{"tusk", "--version"}
	status, err := run(args)
	if err != nil {
		t.Fatal(err)
	}

	expect := "dev\n"
	output := stdout.String()
	if output != expect {
		t.Errorf("want version %q, got %q", expect, output)
	}

	if status != 0 {
		t.Errorf("want exit status 0, got %d", status)
	}
}

func withCapturedOutput() (stdout, stderr *bytes.Buffer, cleanup func()) {
	cleanup = func() {
		ui.LoggerStdout.SetOutput(os.Stdout)
		ui.LoggerStderr.SetOutput(os.Stderr)
	}

	stdout = &bytes.Buffer{}
	stderr = &bytes.Buffer{}

	ui.LoggerStdout.SetOutput(stdout)
	ui.LoggerStderr.SetOutput(stderr)

	return stdout, stderr, cleanup
}
