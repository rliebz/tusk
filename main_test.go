package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"text/template"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"

	"github.com/rliebz/tusk/ui"
)

func Test_run_printVersion(t *testing.T) {
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
	g.Should(be.Equal(stdout.String(), want))
	g.Should(be.Equal(status, 0))
}

func Test_run_printHelp(t *testing.T) {
	tests := []struct {
		args     []string
		wantTmpl string
	}{
		{
			args: []string{"--help"},
			//nolint:lll
			wantTmpl: `{{.}} - the modern task runner

Usage:
   {{.}} [global options] <task> [task options]

Tasks:
   hello                
   lint                 Run static analysis
   print-passed-values  Print values passed

Global Options:
   -f, --file <file>                   Set file to use as the config file
   -h, --help                          Show help and exit
       --install-completion <shell>    Install tab completion for a shell (one of: bash, fish, zsh)
   -q, --quiet                         Only print command output and application errors
   -s, --silent                        Print no output
       --uninstall-completion <shell>  Uninstall tab completion for a shell (one of: bash, fish, zsh)
   -V, --version                       Print version and exit
   -v, --verbose                       Print verbose output
`,
		},
		{
			args: []string{"hello", "--help"},
			wantTmpl: `{{.}} hello

Usage:
   {{.}} hello
`,
		},
		{
			args: []string{"lint", "--help"},
			wantTmpl: `{{.}} lint - Run static analysis

Usage:
   {{.}} lint [options]

Options:
   --fast     Only run fast linters
   --verbose  Run in verbose mode
`,
		},
		{
			args: []string{"print-passed-values", "--help"},
			wantTmpl: `{{.}} print-passed-values - Print values passed

Usage:
   {{.}} print-passed-values [options] <short> <longer-name> <no-details> <values-only>

Description:
   This is a much longer description, which should describe what the task
   does across multiple lines. It rolls over at least two separate times on
   purpose.

Arguments:
   short        The first argument
   longer-name  The second argument
                which is multi-line
                One of: foo, bar
   no-details
   values-only  One of: baz, qux

Options:
       --bool-default-true        Boolean value (default: true)
   -b, --brief                    A brief flag
       --much-less-brief <value>  A much less brief flag
                                  which is multi-line
                                  One of: baz, qux
       --numeric <value>          This is numeric (default: 0)
       --only-default <value>     Default: some-default
       --only-values <value>      One of: alice, bob, carol
       --option-without-usage
       --placeholder <val>        With a value named val
       --usage-default <value>    This is the flag usage (default: 15.5)
       --values-default <value>   Default: alice
                                  One of: alice, bob, carol
`,
		},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			g := ghost.New(t)

			stdout := new(bytes.Buffer)

			status := run(
				config{
					args: slices.Concat([]string{
						"tusk",
						"--file",
						"testdata/help.yml",
					}, tt.args),
					stdout: stdout,
				},
			)

			executable := filepath.Base(os.Args[0])

			tpl := template.Must(template.New("help").Parse(tt.wantTmpl))
			var buf bytes.Buffer
			err := tpl.Execute(&buf, executable)
			g.NoError(err)

			want := buf.String()
			g.Should(be.Equal(stdout.String(), want))
			g.Should(be.Equal(status, 0))
		})
	}
}

func Test_run_exitCodeZero(t *testing.T) {
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
	g.Should(be.Equal(stderr.String(), want))
	g.Should(be.Equal(status, 0))
}

func Test_run_exitCodeNonZero(t *testing.T) {
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

	g.Should(be.Equal(stderr.String(), want))
	g.Should(be.Equal(status, 5))
}

func Test_run_incorrect_usage(t *testing.T) {
	g := ghost.New(t)

	stderr := new(bytes.Buffer)

	args := []string{"tusk", "-f", "./testdata/tusk.yml", "fake-command"}
	status := run(
		config{
			args:   args,
			stderr: stderr,
		},
	)

	wantErr := "Error: task \"fake-command\" is not defined\n"
	g.Should(be.Equal(stderr.String(), wantErr))
	g.Should(be.Equal(status, 1))
}

func Test_run_missing_config(t *testing.T) {
	g := ghost.New(t)

	t.Cleanup(func() { ui.Stdout, ui.Stderr = os.Stdout, os.Stderr })

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	ui.Stdout, ui.Stderr = stdout, stderr

	args := []string{"tusk", "-f", "./testdata/does-not-exist.yml"}
	status := run(
		config{
			args:   args,
			stderr: stderr,
		},
	)

	wantErr := `Error: reading config file "./testdata/does-not-exist.yml": ` +
		"open ./testdata/does-not-exist.yml: no such file or directory\n"

	g.Should(be.Zero(stdout.String()))
	g.Should(be.Equal(stderr.String(), wantErr))
	g.Should(be.Equal(status, 1))
}

func Test_run_completion(t *testing.T) {
	t.Run("unknown task", func(t *testing.T) {
		g := ghost.New(t)

		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)

		args := []string{
			"tusk",
			"-f",
			"./testdata/tusk.yml",
			"fake-command",
			"--generate-bash-completion",
		}
		status := run(config{
			args:   args,
			stdout: stdout,
			stderr: stderr,
		})

		g.Should(be.Equal(status, 0))
		// Once a task has been specified, only task-specific args can be specified
		g.Should(be.Zero(stdout.String()))
		g.Should(be.Zero(stderr.String()))
	})

	t.Run("bad config", func(t *testing.T) {
		g := ghost.New(t)

		t.Cleanup(func() { ui.Stdout, ui.Stderr = os.Stdout, os.Stderr })

		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)
		ui.Stdout, ui.Stderr = stdout, stderr

		args := []string{
			"tusk",
			"-f",
			"./testdata/bad.yml",
			"--generate-bash-completion",
		}
		status := run(config{
			args:   args,
			stdout: stdout,
			stderr: stderr,
		})

		g.Should(be.Equal(status, 0))
		// If we can't parse the config file, we can still show global flags
		g.Should(be.Equal(stdout.String(), `normal
--help:Show help and exit
--install-completion:Install tab completion for a shell (one of: bash, fish, zsh)
--quiet:Only print command output and application errors
--silent:Print no output
--uninstall-completion:Uninstall tab completion for a shell (one of: bash, fish, zsh)
--version:Print version and exit
--verbose:Print verbose output
`))
		g.Should(be.Zero(stderr.String()))
	})

	t.Run("missing config", func(t *testing.T) {
		g := ghost.New(t)

		t.Cleanup(func() { ui.Stdout, ui.Stderr = os.Stdout, os.Stderr })

		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)
		ui.Stdout, ui.Stderr = stdout, stderr

		args := []string{
			"tusk",
			"-f",
			"./testdata/does-not-exist.yml",
			"--generate-bash-completion",
		}
		status := run(config{
			args:   args,
			stdout: stdout,
			stderr: stderr,
		})

		g.Should(be.Equal(status, 0))
		// If we can't parse the config file, we can still show global flags
		g.Should(be.Equal(stdout.String(), `normal
--help:Show help and exit
--install-completion:Install tab completion for a shell (one of: bash, fish, zsh)
--quiet:Only print command output and application errors
--silent:Print no output
--uninstall-completion:Uninstall tab completion for a shell (one of: bash, fish, zsh)
--version:Print version and exit
--verbose:Print verbose output
`))
		g.Should(be.Zero(stderr.String()))
	})
}
