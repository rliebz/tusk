package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestRun_printVersion(t *testing.T) {
	registerCleanup(t)
	stdout := new(bytes.Buffer)

	args := []string{"tusk", "--version"}
	status := run(
		config{
			args:   args,
			stdout: stdout,
		},
	)

	want := "dev\n"
	got := stdout.String()
	assert.Check(t, cmp.Equal(want, got))
	assert.Check(t, status == 0)
}

func TestRun_printHelp(t *testing.T) {
	registerCleanup(t)
	stdout := new(bytes.Buffer)

	args := []string{"tusk", "--help"}
	status := run(
		config{
			args:   args,
			stdout: stdout,
		},
	)

	executable := filepath.Base(os.Args[0])

	message := `{{.}} - the modern task runner

Usage:
   {{.}} [global options] <task> [task options]

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

	tpl := template.Must(template.New("help").Parse(message))
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, executable); err != nil {
		t.Fatal(err)
	}

	want := buf.String()
	assert.Check(t, cmp.Equal(want, stdout.String()))
	assert.Check(t, status == 0)
}

func TestRun_exitCodeZero(t *testing.T) {
	registerCleanup(t)
	stderr := new(bytes.Buffer)

	args := []string{"tusk", "-f", "./testdata/tusk.yml", "exit", "0"}
	status := run(
		config{
			args:   args,
			stderr: stderr,
		},
	)
	assert.Check(t, cmp.Equal(status, 0))

	want := "exit $ exit 0\n"
	assert.Check(t, cmp.Equal(want, stderr.String()))
}

func TestRun_exitCodeNonZero(t *testing.T) {
	registerCleanup(t)
	stderr := new(bytes.Buffer)

	args := []string{"tusk", "-f", "./testdata/tusk.yml", "exit", "5"}
	status := run(
		config{
			args:   args,
			stderr: stderr,
		},
	)
	assert.Check(t, cmp.Equal(status, 5))

	want := `exit $ exit 5
exit status 5
`

	assert.Check(t, cmp.Equal(want, stderr.String()))
}

func TestRun_incorrectUsage(t *testing.T) {
	registerCleanup(t)
	stderr := new(bytes.Buffer)

	args := []string{"tusk", "-f", "./testdata/tusk.yml", "fake-command"}
	status := run(
		config{
			args:   args,
			stderr: stderr,
		},
	)
	assert.Check(t, cmp.Equal(status, 1))

	want := "Error: No help topic for 'fake-command'\n"
	assert.Equal(t, want, stderr.String())
}

// registerCleanup is needed because main calls os.Chdir.
//
// Ideally, we never actually call os.Chdir, and instead set each command's
// working directory and resolve relative file paths manually.
func registerCleanup(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Error(err)
		}
	})
}
