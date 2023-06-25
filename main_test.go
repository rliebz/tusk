package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
)

func TestRun_printVersion(t *testing.T) {
	g := ghost.New(t)

	stdout := new(bytes.Buffer)

	args := []string{"tusk", "--version"}
	status := run(
		config{
			args:   args,
			stdout: stdout,
		},
	)

	want := "(devel)\n"
	g.Should(be.Equal(want, stdout.String()))
	g.Should(be.Equal(0, status))
}

func TestRun_printHelp(t *testing.T) {
	g := ghost.New(t)

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
   lint     Run static analysis
   release  Release the latest version with goreleaser
   test     Run the tests
   tidy     Clean up and format the repo

Global Options:
   -f, --file <file>                   Set file to use as the config file
   -h, --help                          Show help and exit
       --install-completion <shell>    Install tab completion for a shell
   -q, --quiet                         Only print command output and application errors
   -s, --silent                        Print no output
       --uninstall-completion <shell>  Uninstall tab completion for a shell
   -V, --version                       Print version and exit
   -v, --verbose                       Print verbose output
`

	tpl := template.Must(template.New("help").Parse(message))
	var buf bytes.Buffer
	err := tpl.Execute(&buf, executable)
	g.NoError(err)

	want := buf.String()
	g.Should(be.Equal(want, stdout.String()))
	g.Should(be.Equal(0, status))
}

func TestRun_exitCodeZero(t *testing.T) {
	g := ghost.New(t)

	stderr := new(bytes.Buffer)

	args := []string{"tusk", "-f", "./testdata/tusk.yml", "exit", "0"}
	status := run(
		config{
			args:   args,
			stderr: stderr,
		},
	)

	want := "exit $ exit 0\n"
	g.Should(be.Equal(want, stderr.String()))
	g.Should(be.Equal(0, status))
}

func TestRun_exitCodeNonZero(t *testing.T) {
	g := ghost.New(t)

	stderr := new(bytes.Buffer)

	args := []string{"tusk", "-f", "./testdata/tusk.yml", "exit", "5"}
	status := run(
		config{
			args:   args,
			stderr: stderr,
		},
	)

	want := `exit $ exit 5
exit status 5
`

	g.Should(be.Equal(want, stderr.String()))
	g.Should(be.Equal(5, status))
}

func TestRun_incorrectUsage(t *testing.T) {
	g := ghost.New(t)

	stderr := new(bytes.Buffer)

	args := []string{"tusk", "-f", "./testdata/tusk.yml", "fake-command"}
	status := run(
		config{
			args:   args,
			stderr: stderr,
		},
	)

	want := "Error: No help topic for 'fake-command'\n"
	g.Should(be.Equal(want, stderr.String()))
	g.Should(be.Equal(1, status))
}
