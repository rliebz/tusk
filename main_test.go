package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/rliebz/tusk/ui"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestRun_printVersion(t *testing.T) {
	stdout, _, cleanup := setupTestSandbox(t)
	defer cleanup()

	args := []string{"tusk", "--version"}
	status, err := run(args)
	assert.NilError(t, err)

	want := "dev\n"
	got := stdout.String()
	assert.Check(t, cmp.Equal(want, got))
	assert.Check(t, status == 0)
}

func TestRun_printHelp(t *testing.T) {
	stdout, _, cleanup := setupTestSandbox(t)
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

func TestRun_exitCodeZero(t *testing.T) {
	_, stderr, cleanup := setupTestSandbox(t)
	defer cleanup()

	args := []string{"tusk", "-f", "./testdata/tusk.yml", "exit", "0"}
	status, err := run(args)
	assert.NilError(t, err)

	want := `exit $ exit 0
`

	got := stderr.String()

	assert.Check(t, cmp.Equal(want, got))
	assert.Check(t, cmp.Equal(status, 0))
}

func TestRun_exitCodeNonZero(t *testing.T) {
	_, stderr, cleanup := setupTestSandbox(t)
	defer cleanup()

	args := []string{"tusk", "-f", "./testdata/tusk.yml", "exit", "5"}
	status, err := run(args)
	assert.NilError(t, err)

	want := `exit $ exit 5
exit status 5
`

	got := stderr.String()

	assert.Check(t, cmp.Equal(want, got))
	assert.Check(t, cmp.Equal(status, 5))
}

func TestRun_incorrectUsage(t *testing.T) {
	_, _, cleanup := setupTestSandbox(t)
	defer cleanup()

	args := []string{"tusk", "-f", "./testdata/tusk.yml", "fake-command"}
	status, err := run(args)
	assert.Error(t, err, "No help topic for 'fake-command'")

	assert.Check(t, cmp.Equal(status, 1))
}

func TestRun_badTuskYml(t *testing.T) {
	_, _, cleanup := setupTestSandbox(t)
	defer cleanup()

	args := []string{"tusk", "-f", "./testdata/bad.yml"}
	status, err := run(args)
	assert.ErrorContains(t, err, "field key not found")

	assert.Check(t, cmp.Equal(status, 1))
}

func setupTestSandbox(t *testing.T) (stdout, stderr *bytes.Buffer, cleanup func()) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cleanup = func() {
		ui.LoggerStdout.SetOutput(os.Stdout)
		ui.LoggerStderr.SetOutput(os.Stderr)
		if err := os.Chdir(wd); err != nil {
			t.Error(err)
		}
	}

	stdout = &bytes.Buffer{}
	stderr = &bytes.Buffer{}

	ui.LoggerStdout.SetOutput(stdout)
	ui.LoggerStderr.SetOutput(stderr)

	return stdout, stderr, cleanup
}
