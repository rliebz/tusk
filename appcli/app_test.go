package appcli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	"github.com/urfave/cli"

	"github.com/rliebz/tusk/runner"
	"github.com/rliebz/tusk/ui"
)

func TestNewFlagApp(t *testing.T) {
	g := ghost.New(t)

	cfgText := []byte(`options:
  foo:
    short: f
    default: foovalue

tasks:
  mytask:
    run: echo ${foo}
`)

	flagApp, err := newMetaApp(cfgText)
	g.NoError(err)

	err = flagApp.Run([]string{"tusk", "mytask", "--foo", "other"})
	g.NoError(err)

	command, ok := flagApp.Metadata["command"].(*cli.Command)
	g.Must(be.True(ok))

	g.Should(be.Equal(command.Name, "mytask"))

	flags, ok := flagApp.Metadata["flagsPassed"].(map[string]string)
	g.Must(be.True(ok))

	g.Should(be.DeepEqual(flags, map[string]string{"foo": "other"}))
}

func TestNewFlagApp_no_options(t *testing.T) {
	g := ghost.New(t)

	cfgText := []byte(`tasks:
  mytask:
    run: echo foo
`)

	flagApp, err := newMetaApp(cfgText)
	g.NoError(err)

	err = flagApp.Run([]string{"tusk", "mytask"})
	g.NoError(err)

	command, ok := flagApp.Metadata["command"].(*cli.Command)
	g.Must(be.True(ok))

	g.Should(be.Equal(command.Name, "mytask"))

	flags, ok := flagApp.Metadata["flagsPassed"].(map[string]string)
	g.Must(be.True(ok))

	g.Should(be.DeepEqual(flags, map[string]string{}))
}

func TestNewApp(t *testing.T) {
	g := ghost.New(t)

	taskName := "foo"
	name := "new-name"
	usage := "new usage"

	cfgText := fmt.Appendf(nil, `
name: %s
usage: %s
tasks: { %q: {} }
`,
		name, usage, taskName,
	)
	meta := &runner.Metadata{
		CfgText: cfgText,
		Logger:  ui.Noop(),
	}

	app, err := NewApp([]string{"tusk", taskName}, meta)
	g.NoError(err)

	g.Should(be.SliceLen(app.Commands, 1))
	g.Should(be.Equal(app.Name, name))
	g.Should(be.Equal(app.Usage, usage))
}

func TestNewApp_exit_code(t *testing.T) {
	g := ghost.New(t)

	args := []string{"tusk", "foo"}
	wantExitCode := 99
	cfgText := []byte(`
tasks:
  foo:
    run: exit 99`)
	meta := &runner.Metadata{
		CfgText: cfgText,
		Logger:  ui.Noop(),
	}

	app, err := NewApp(args, meta)
	g.NoError(err)

	g.Must(be.SliceLen(app.Commands, 1))

	err = app.Run(args)
	var exitErr *exec.ExitError
	ok := errors.As(err, &exitErr)
	g.Must(be.True(ok))

	exitCode := exitErr.Sys().(syscall.WaitStatus).ExitStatus()
	g.Should(be.Equal(exitCode, wantExitCode))
}

func TestNewApp_print_help(t *testing.T) {
	g := ghost.New(t)

	args := []string{"tusk"}
	cfgText := []byte(`
tasks:
  my-task:
    run: exit 99`)
	meta := &runner.Metadata{
		CfgText: cfgText,
		Logger:  ui.Noop(),
	}

	app, err := NewApp(args, meta)
	g.NoError(err)

	var buf bytes.Buffer
	app.Writer = &buf

	err = app.Run(args)
	g.NoError(err)

	g.Should(be.StringMatching(
		buf.String(),
		`^appcli\.test.* - the modern task runner`,
	))
}

func TestNewApp_task_not_found(t *testing.T) {
	g := ghost.New(t)

	args := []string{"tusk", "fake-task"}
	cfgText := []byte(`
tasks:
  my-task:
    run: exit 99`)
	meta := &runner.Metadata{
		CfgText: cfgText,
		Logger:  ui.Noop(),
	}

	_, err := NewApp(args, meta)
	g.Should(be.ErrorEqual(err, `task "fake-task" is not defined`))
}

func TestNewApp_private_task(t *testing.T) {
	g := ghost.New(t)

	args := []string{"tusk", "public"}
	wantExitCode := 99
	cfgText := []byte(`
tasks:
  private:
    private: true
    run: exit 99
  public:
    run: {task: private}`)
	meta := &runner.Metadata{
		CfgText: cfgText,
		Logger:  ui.Noop(),
	}

	app, err := NewApp(args, meta)
	g.NoError(err)

	g.Must(be.SliceLen(app.Commands, 1))

	// Ensure private task still runs as subtask
	err = app.Run(args)
	var exitErr *exec.ExitError
	ok := errors.As(err, &exitErr)
	g.Must(be.True(ok))

	exitCode := exitErr.Sys().(syscall.WaitStatus).ExitStatus()
	g.Should(be.Equal(exitCode, wantExitCode))
}

func TestNewApp_bad_config(t *testing.T) {
	g := ghost.New(t)

	_, err := NewApp(
		[]string{"tusk"},
		&runner.Metadata{CfgText: []byte(`invalid`)},
	)
	g.Should(be.ErrorEqual(
		err,
		"yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `invalid` into runner.configType",
	))
}

func TestNewApp_bad_flag(t *testing.T) {
	g := ghost.New(t)

	_, err := NewApp([]string{"tusk", "--invalid"}, &runner.Metadata{})
	g.Should(be.ErrorEqual(err, "flag provided but not defined: -invalid"))
}

func TestGetConfigMetadata_defaults(t *testing.T) {
	g := ghost.New(t)

	args := []string{"tusk"}

	meta, err := NewMetadata(ui.Noop(), args)
	g.NoError(err)

	// The project's tuskfile should be found in the project root.
	wd, err := os.Getwd()
	g.NoError(err)

	g.Should(be.Equal(meta.CfgPath, filepath.Join(filepath.Dir(wd), "tusk.yml")))
	g.Should(be.Equal(meta.Logger.Level(), ui.LevelNormal))
	g.Should(be.False(meta.PrintVersion))
}

func TestGetConfigMetadata_file(t *testing.T) {
	g := ghost.New(t)

	cfgPath := "testdata/example.yml"
	args := []string{"tusk", "--file", cfgPath}

	meta, err := NewMetadata(ui.Noop(), args)
	g.NoError(err)

	g.Should(be.Equal(meta.CfgPath, cfgPath))

	cfgText, err := os.ReadFile(cfgPath)
	g.NoError(err)

	g.Should(be.Equal(string(meta.CfgText), string(cfgText)))
}

func TestGetConfigMetadata_fileNoExist(t *testing.T) {
	g := ghost.New(t)

	_, err := NewMetadata(ui.Noop(), []string{"tusk", "--file", "fakefile.yml"})
	if !g.Should(be.True(errors.Is(err, os.ErrNotExist))) {
		t.Log(err)
	}
}

func TestGetConfigMetadata_version(t *testing.T) {
	g := ghost.New(t)

	meta, err := NewMetadata(ui.Noop(), []string{"tusk", "--version"})
	g.NoError(err)

	g.Should(be.True(meta.PrintVersion))
}

func TestGetConfigMetadata_verbosity(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want ui.Level
	}{
		{
			"normal",
			[]string{"tusk"},
			ui.LevelNormal,
		},
		{
			"silent",
			[]string{"tusk", "--silent"},
			ui.LevelSilent,
		},
		{
			"quiet",
			[]string{"tusk", "--quiet"},
			ui.LevelQuiet,
		},
		{
			"verbose",
			[]string{"tusk", "--verbose"},
			ui.LevelVerbose,
		},
		{
			"quiet verbose",
			[]string{"tusk", "--quiet", "--verbose"},
			ui.LevelQuiet,
		},
		{
			"silent quiet",
			[]string{"tusk", "--silent", "--quiet"},
			ui.LevelSilent,
		},
		{
			"silent verbose",
			[]string{"tusk", "--silent", "--verbose"},
			ui.LevelSilent,
		},
		{
			"silent quiet verbose",
			[]string{"tusk", "--silent", "--quiet", "--verbose"},
			ui.LevelSilent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			meta, err := NewMetadata(ui.Noop(), tt.args)
			g.NoError(err)

			g.Should(be.Equal(meta.Logger.Level(), tt.want))
		})
	}
}
