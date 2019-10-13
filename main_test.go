package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/rliebz/tusk/ui"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestRun_PrintVersion(t *testing.T) {
	stdout, _, cleanup := withCapturedOutput()
	defer cleanup()

	args := []string{"tusk", "--version"}
	status, err := run(args)
	assert.NilError(t, err)

	want := "dev\n"
	got := stdout.String()
	assert.Check(t, cmp.Equal(want, got))
	assert.Check(t, status == 0)
}

func TestRun_PrintHelp(t *testing.T) {
	stdout, _, cleanup := withCapturedOutput()
	defer cleanup()

	args := []string{"tusk", "--help"}
	status, err := run(args)
	assert.NilError(t, err)

	want := `tusk.test - the modern task runner

Usage:
   tusk.test [global options] <task> [task options] 

Tasks:
   bootstrap  Set up app dependencies for first time use
   circleci   Run the circleci build locally
   lint       Run static analysis
   release    Release the latest version with goreleaser
   test       Run the tests
   tidy       Clean up and format the repo

Global Options:
   -f, --file <file>  Set file to use as the config file
   -h, --help         Show help and exit
   -q, --quiet        Only print command output and application errors
   -s, --silent       Print no output
   -V, --version      Print version and exit
   -v, --verbose      Print verbose output
`

	got := stdout.String()
	assert.Check(t, cmp.Equal(want, got))
	assert.Check(t, status == 0)
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
