package appcli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/tusk/runner"
	"github.com/rliebz/tusk/ui"
	"github.com/urfave/cli"
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
	g.Must(ghost.BeTrue(ok))

	g.Should(ghost.Equal("mytask", command.Name))

	flags, ok := flagApp.Metadata["flagsPassed"].(map[string]string)
	g.Must(ghost.BeTrue(ok))

	g.Should(ghost.DeepEqual(map[string]string{"foo": "other"}, flags))
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
	g.Must(ghost.BeTrue(ok))

	g.Should(ghost.Equal("mytask", command.Name))

	flags, ok := flagApp.Metadata["flagsPassed"].(map[string]string)
	g.Must(ghost.BeTrue(ok))

	g.Should(ghost.DeepEqual(map[string]string{}, flags))
}

func TestNewApp(t *testing.T) {
	g := ghost.New(t)

	taskName := "foo"
	name := "new-name"
	usage := "new usage"

	cfgText := []byte(fmt.Sprintf(`
name: %s
usage: %s
tasks: { %q: {} }
`,
		name, usage, taskName,
	))
	meta := &runner.Metadata{CfgText: cfgText}

	app, err := NewApp([]string{"tusk", taskName}, meta)
	g.NoError(err)

	g.Should(ghost.Len(1, app.Commands))
	g.Should(ghost.Equal(name, app.Name))
	g.Should(ghost.Equal(usage, app.Usage))
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

	g.Must(ghost.Len(1, app.Commands))

	err = app.Run(args)
	var exitErr *exec.ExitError
	ok := errors.As(err, &exitErr)
	g.Must(ghost.BeTrue(ok))

	exitCode := exitErr.Sys().(syscall.WaitStatus).ExitStatus()
	g.Should(ghost.Equal(wantExitCode, exitCode))
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

	g.Must(ghost.Len(1, app.Commands))

	// Ensure private task still runs as subtask
	err = app.Run(args)
	var exitErr *exec.ExitError
	ok := errors.As(err, &exitErr)
	g.Must(ghost.BeTrue(ok))

	exitCode := exitErr.Sys().(syscall.WaitStatus).ExitStatus()
	g.Should(ghost.Equal(wantExitCode, exitCode))
}

func TestNewApp_fails_bad_config(t *testing.T) {
	g := ghost.New(t)

	_, err := NewApp(
		[]string{"tusk"},
		&runner.Metadata{CfgText: []byte(`invalid`)},
	)
	g.Should(ghost.ErrorEqual(err,
		"yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `invalid` into runner.configType",
	))
}

func TestNewApp_fails_bad_flag(t *testing.T) {
	g := ghost.New(t)

	_, err := NewApp([]string{"tusk", "--invalid"}, &runner.Metadata{})
	g.Should(ghost.ErrorEqual(err, "flag provided but not defined: -invalid"))
}

func TestGetConfigMetadata_defaults(t *testing.T) {
	g := ghost.New(t)

	args := []string{"tusk"}

	metadata, err := GetConfigMetadata(args)
	g.NoError(err)

	// The project's tuskfile should be found in the project root.
	wd, err := os.Getwd()
	g.NoError(err)

	g.Should(ghost.Equal(filepath.Dir(wd), metadata.Directory))
	g.Should(ghost.Equal(ui.VerbosityLevelNormal, metadata.Logger.Verbosity))
	g.ShouldNot(ghost.BeTrue(metadata.PrintVersion))
}

func TestGetConfigMetadata_file(t *testing.T) {
	g := ghost.New(t)

	cfgPath := "testdata/example.yml"
	args := []string{"tusk", "--file", cfgPath}

	metadata, err := GetConfigMetadata(args)
	g.NoError(err)

	g.Should(ghost.Equal("testdata", metadata.Directory))

	cfgText, err := os.ReadFile(cfgPath)
	g.NoError(err)

	g.Should(ghost.Equal(string(cfgText), string(metadata.CfgText)))
}

func TestGetConfigMetadata_fileNoExist(t *testing.T) {
	g := ghost.New(t)

	_, err := GetConfigMetadata([]string{"tusk", "--file", "fakefile.yml"})
	if !g.Should(ghost.BeTrue(errors.Is(err, os.ErrNotExist))) {
		t.Log(err)
	}
}

func TestGetConfigMetadata_version(t *testing.T) {
	g := ghost.New(t)

	metadata, err := GetConfigMetadata([]string{"tusk", "--version"})
	g.NoError(err)

	g.Should(ghost.BeTrue(metadata.PrintVersion))
}

func TestGetConfigMetadata_verbosity(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want ui.VerbosityLevel
	}{
		{
			"normal",
			[]string{"tusk"},
			ui.VerbosityLevelNormal,
		},
		{
			"silent",
			[]string{"tusk", "--silent"},
			ui.VerbosityLevelSilent,
		},
		{
			"quiet",
			[]string{"tusk", "--quiet"},
			ui.VerbosityLevelQuiet,
		},
		{
			"verbose",
			[]string{"tusk", "--verbose"},
			ui.VerbosityLevelVerbose,
		},
		{
			"quiet verbose",
			[]string{"tusk", "--quiet", "--verbose"},
			ui.VerbosityLevelQuiet,
		},
		{
			"silent quiet",
			[]string{"tusk", "--silent", "--quiet"},
			ui.VerbosityLevelSilent,
		},
		{
			"silent verbose",
			[]string{"tusk", "--silent", "--verbose"},
			ui.VerbosityLevelSilent,
		},
		{
			"silent quiet verbose",
			[]string{"tusk", "--silent", "--quiet", "--verbose"},
			ui.VerbosityLevelSilent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			metadata, err := GetConfigMetadata(tt.args)
			g.NoError(err)

			g.Should(ghost.Equal(tt.want, metadata.Logger.Verbosity))
		})
	}
}
