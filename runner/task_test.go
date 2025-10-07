package runner

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	yaml "gopkg.in/yaml.v2"

	"github.com/rliebz/tusk/internal/xtesting"
	"github.com/rliebz/tusk/marshal"
	"github.com/rliebz/tusk/ui"
)

func TestTask_UnmarshalYAML(t *testing.T) {
	g := ghost.New(t)

	wd, err := os.Getwd()
	g.NoError(err)

	testdata := func(filename string) string {
		return filepath.Join(wd, "testdata", filename)
	}

	tests := []struct {
		name    string
		input   string
		want    Task
		wantErr string
	}{
		{
			name: "options and args",
			input: `
options: { one: {}, two: {} }
args: { three: {}, four: {} }
`,
			want: Task{
				Options: Options{
					{
						Passable: Passable{
							Name: "one",
						},
					},
					{
						Passable: Passable{
							Name: "two",
						},
					},
				},
				Args: Args{
					{
						Passable: Passable{
							Name: "three",
						},
					},
					{
						Passable: Passable{
							Name: "four",
						},
					},
				},
			},
		},
		{
			name:  "include",
			input: fmt.Sprintf(`{include: %q}`, testdata("included.yml")),
			want: Task{
				Usage: "A valid example of an included task",
				RunList: marshal.Slice[*Run]{{Command: marshal.Slice[*Command]{{
					Exec:  `echo "We're in!"`,
					Print: `echo "We're in!"`,
				}}}},
			},
		},
		{
			name:    "include-extra",
			input:   fmt.Sprintf(`{include: %q, usage: "This is incorrect"}`, testdata("included.yml")),
			wantErr: `tasks using "include" may not specify other fields`,
		},
		{
			name:    "include invalid",
			input:   fmt.Sprintf(`{include: %q}`, testdata("included-invalid.yml")),
			wantErr: "decoding included file",
		},
		{
			name:    "include missing",
			input:   fmt.Sprintf(`{include: %q}`, testdata("not-a-real-file.yml")),
			wantErr: "opening included file",
		},
		{
			name:    "invalid",
			input:   "[invalid]",
			wantErr: "yaml: unmarshal errors",
		},
		{
			name: "option and arg share name",
			input: `
options: { foo: {} }
args: { foo: {} }
`,
			wantErr: `argument and option "foo" must have unique names within a task`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			var got Task
			err := yaml.UnmarshalStrict([]byte(tt.input), &got)
			if tt.wantErr != "" {
				g.Should(be.ErrorContaining(err, tt.wantErr))
				return
			}
			g.NoError(err)

			g.Should(be.DeepEqual(got, tt.want))
		})
	}
}

func TestTask_Execute_errors_returned(t *testing.T) {
	tests := []struct {
		name    string
		run     string
		finally string
		wantErr string
	}{
		{
			name:    "run error only",
			run:     "exit 1",
			finally: "exit 0",
			wantErr: "exit status 1",
		},
		{
			name:    "finally error only",
			run:     "exit 0",
			finally: "exit 1",
			wantErr: "exit status 1",
		},
		{
			name:    "run and finally error",
			run:     "exit 1",
			finally: "exit 2",
			wantErr: "exit status 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			run := Run{Command: marshal.Slice[*Command]{{Exec: tt.run}}}
			finally := Run{Command: marshal.Slice[*Command]{{Exec: tt.finally}}}
			task := Task{
				RunList: marshal.Slice[*Run]{&run},
				Finally: marshal.Slice[*Run]{&finally},
			}

			err := task.Execute(Context{Logger: ui.Noop()})
			g.Should(be.ErrorEqual(err, tt.wantErr))
		})
	}
}

func TestTask_Execute_cache(t *testing.T) {
	tests := []struct {
		name          string
		skipWindows   bool
		source        marshal.Slice[string]
		target        marshal.Slice[string]
		mutate        func(t *testing.T)
		wantRunCount  int
		wantFirstErr  string
		wantSecondErr string
	}{
		{
			name: "one source one target",
			source: marshal.Slice[string]{
				"a1.txt",
			},
			target: marshal.Slice[string]{
				"b1.txt",
			},
			wantRunCount: 1,
		},
		{
			name: "multi source multi target",
			source: marshal.Slice[string]{
				"a1.txt",
				"b1.txt",
			},
			target: marshal.Slice[string]{
				"c2.txt",
				"d2.txt",
			},
			wantRunCount: 1,
		},
		{
			name: "removed source",
			source: marshal.Slice[string]{
				"a1.txt",
				"b1.txt",
			},
			target: marshal.Slice[string]{
				"c2.txt",
				"d2.txt",
			},
			mutate: func(t *testing.T) {
				g := ghost.New(t)
				err := os.Remove("a1.txt")
				g.NoError(err)
			},
			wantSecondErr: "no source files found matching pattern: a1.txt",
		},
		{
			name: "removed target",
			source: marshal.Slice[string]{
				"a1.txt",
				"b1.txt",
			},
			target: marshal.Slice[string]{
				"c2.txt",
				"d2.txt",
			},
			mutate: func(t *testing.T) {
				g := ghost.New(t)
				err := os.Remove("c2.txt")
				g.NoError(err)
			},
			wantRunCount: 2,
		},
		{
			name: "modified source",
			source: marshal.Slice[string]{
				"a1.txt",
				"b1.txt",
			},
			target: marshal.Slice[string]{
				"c2.txt",
				"d2.txt",
			},
			mutate: func(t *testing.T) {
				g := ghost.New(t)
				err := os.WriteFile("a1.txt", []byte("different data"), 0o600)
				g.NoError(err)
			},
			wantRunCount: 2,
		},
		{
			name: "modified target",
			source: marshal.Slice[string]{
				"a1.txt",
				"b1.txt",
			},
			target: marshal.Slice[string]{
				"c2.txt",
				"d2.txt",
			},
			mutate: func(t *testing.T) {
				g := ghost.New(t)
				err := os.WriteFile("c2.txt", []byte("different data"), 0o600)
				g.NoError(err)
			},
			wantRunCount: 2,
		},
		{
			name: "glob no change",
			source: marshal.Slice[string]{
				"*1.txt",
			},
			target: marshal.Slice[string]{
				"*2.txt",
			},
			wantRunCount: 1,
		},
		{
			name: "glob modified source",
			source: marshal.Slice[string]{
				"*1.txt",
			},
			target: marshal.Slice[string]{
				"*2.txt",
			},
			mutate: func(t *testing.T) {
				g := ghost.New(t)
				err := os.WriteFile("a1.txt", []byte("different data"), 0o600)
				g.NoError(err)
			},
			wantRunCount: 2,
		},
		{
			name: "glob modified target",
			source: marshal.Slice[string]{
				"*1.txt",
			},
			target: marshal.Slice[string]{
				"*2.txt",
			},
			mutate: func(t *testing.T) {
				g := ghost.New(t)
				err := os.WriteFile("c2.txt", []byte("different data"), 0o600)
				g.NoError(err)
			},
			wantRunCount: 2,
		},
		{
			name: "glob new source",
			source: marshal.Slice[string]{
				"*1.txt",
			},
			target: marshal.Slice[string]{
				"*2.txt",
			},
			mutate: func(t *testing.T) {
				g := ghost.New(t)
				err := os.WriteFile("x1.txt", []byte("different data"), 0o600)
				g.NoError(err)
			},
			wantRunCount: 2,
		},
		{
			name: "glob new target",
			source: marshal.Slice[string]{
				"*1.txt",
			},
			target: marshal.Slice[string]{
				"*2.txt",
			},
			mutate: func(t *testing.T) {
				g := ghost.New(t)
				err := os.WriteFile("x2.txt", []byte("different data"), 0o600)
				g.NoError(err)
			},
			wantRunCount: 2,
		},
		{
			name: "glob removed source",
			source: marshal.Slice[string]{
				"*1.txt",
			},
			target: marshal.Slice[string]{
				"*2.txt",
			},
			mutate: func(t *testing.T) {
				g := ghost.New(t)
				err := os.Remove("a1.txt")
				g.NoError(err)
			},
			wantRunCount: 2,
		},
		{
			name: "glob removed target",
			source: marshal.Slice[string]{
				"*1.txt",
			},
			target: marshal.Slice[string]{
				"*2.txt",
			},
			mutate: func(t *testing.T) {
				g := ghost.New(t)
				err := os.Remove("c2.txt")
				g.NoError(err)
			},
			wantRunCount: 2,
		},
		{
			name: "glob no sources",
			source: marshal.Slice[string]{
				"*3.txt",
			},
			target: marshal.Slice[string]{
				"*2.txt",
			},
			wantFirstErr: "no source files found matching pattern: *3.txt",
		},
		{
			name: "glob no targets",
			source: marshal.Slice[string]{
				"*1.txt",
			},
			target: marshal.Slice[string]{
				"*3.txt",
			},
			wantRunCount: 2,
		},
		{
			name: "glob partial sources",
			source: marshal.Slice[string]{
				"*1.txt",
				"*3.txt",
			},
			target: marshal.Slice[string]{
				"*2.txt",
			},
			wantFirstErr: "no source files found matching pattern: *3.txt",
		},
		{
			name: "glob partial targets",
			source: marshal.Slice[string]{
				"*1.txt",
			},
			target: marshal.Slice[string]{
				"*2.txt",
				"*3.txt",
			},
			wantRunCount: 2,
		},
		{
			name: "relative paths",
			source: marshal.Slice[string]{
				"./a1.txt",
			},
			target: marshal.Slice[string]{
				"./b1.txt",
			},
			wantRunCount: 1,
		},
		{
			name:        "unreadable source",
			skipWindows: true,
			source: marshal.Slice[string]{
				"writeonly.txt",
			},
			target: marshal.Slice[string]{
				"c2.txt",
			},
			wantFirstErr: "open writeonly.txt: permission denied",
		},
		{
			name:        "unreadable target",
			skipWindows: true,
			source: marshal.Slice[string]{
				"a1.txt",
			},
			target: marshal.Slice[string]{
				"writeonly.txt",
			},
			wantFirstErr: "caching task: open writeonly.txt: permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipWindows && runtime.GOOS == "windows" {
				t.Skip("test is for unix file permissions")
			}

			g := ghost.New(t)

			wd := xtesting.UseTempDir(t)
			cfgPath := filepath.Join(wd, "tusk.yml")

			err := os.WriteFile("a1.txt", []byte("data a"), 0o600)
			g.NoError(err)

			err = os.WriteFile("b1.txt", []byte("data b"), 0o600)
			g.NoError(err)

			err = os.WriteFile("c2.txt", []byte("data c"), 0o600)
			g.NoError(err)

			err = os.WriteFile("d2.txt", []byte("data d"), 0o600)
			g.NoError(err)

			err = os.WriteFile("writeonly.txt", []byte("data writeonly"), 0o200)
			g.NoError(err)

			cacheHome := t.TempDir()
			t.Setenv("XDG_CACHE_HOME", cacheHome)

			var buf bytes.Buffer

			logger := ui.New(ui.Config{
				Stdout: io.Discard,
				Stderr: &buf,
			})

			ctx := Context{
				CfgPath: cfgPath,
				Logger:  logger,
			}

			task := Task{
				Name:   "my-task",
				Source: tt.source,
				Target: tt.target,
				RunList: marshal.Slice[*Run]{
					{
						Command: marshal.Slice[*Command]{
							{
								Print: "exit 0",
								Exec:  "exit 0",
							},
						},
					},
				},
			}

			err = task.Execute(ctx)
			if tt.wantFirstErr != "" {
				g.Should(be.ErrorEqual(err, tt.wantFirstErr))
				return
			}
			g.NoError(err)

			if tt.mutate != nil {
				tt.mutate(t)
			}

			err = task.Execute(ctx)
			if tt.wantSecondErr != "" {
				g.Should(be.ErrorEqual(err, tt.wantSecondErr))
				return
			}
			g.NoError(err)

			runCount := strings.Count(buf.String(), "exit 0")
			if !g.Should(be.Equal(runCount, tt.wantRunCount)) {
				t.Log(buf.String())
			}
		})
	}

	t.Run("readonly cache path", func(t *testing.T) {
		g := ghost.New(t)

		wd := xtesting.UseTempDir(t)
		cfgPath := filepath.Join(wd, "tusk.yml")

		err := os.WriteFile("input.txt", []byte("data a"), 0o600)
		g.NoError(err)

		err = os.WriteFile("output.txt", []byte("data b"), 0o600)
		g.NoError(err)

		cacheHome := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", cacheHome)

		var buf bytes.Buffer

		logger := ui.New(ui.Config{
			Stdout: io.Discard,
			Stderr: &buf,
		})

		ctx := Context{
			CfgPath: cfgPath,
			Logger:  logger,
		}

		task := Task{
			Name: "my-task",
			Source: marshal.Slice[string]{
				"input.txt",
			},
			Target: marshal.Slice[string]{
				"output.txt",
			},
			RunList: marshal.Slice[*Run]{
				{
					Command: marshal.Slice[*Command]{
						{
							Print: "exit 0",
							Exec:  "exit 0",
						},
					},
				},
			},
		}

		cachePath, err := task.taskInputCachePath(ctx)
		g.NoError(err)

		err = os.MkdirAll(filepath.Dir(cachePath), 0o700)
		g.NoError(err)

		err = os.WriteFile(cachePath, []byte("data readonly"), 0o400)
		g.NoError(err)

		permissionDenied := "permission denied"
		if runtime.GOOS == "windows" {
			permissionDenied = "Access is denied."
		}

		err = task.Execute(ctx)
		g.Should(be.ErrorEqual(
			err,
			fmt.Sprintf("caching task: open %s: %s", cachePath, permissionDenied),
		))

		runCount := strings.Count(buf.String(), "exit 0")
		if !g.Should(be.Equal(runCount, 1)) {
			t.Log(buf.String())
		}
	})

	t.Run("writeonly cache path", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("test is for unix file permissions")
		}

		g := ghost.New(t)

		wd := xtesting.UseTempDir(t)
		cfgPath := filepath.Join(wd, "tusk.yml")

		err := os.WriteFile("input.txt", []byte("data a"), 0o600)
		g.NoError(err)

		err = os.WriteFile("output.txt", []byte("data b"), 0o600)
		g.NoError(err)

		cacheHome := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", cacheHome)

		var buf bytes.Buffer

		logger := ui.New(ui.Config{
			Stdout: io.Discard,
			Stderr: &buf,
		})

		ctx := Context{
			CfgPath: cfgPath,
			Logger:  logger,
		}

		task := Task{
			Name: "my-task",
			Source: marshal.Slice[string]{
				"input.txt",
			},
			Target: marshal.Slice[string]{
				"output.txt",
			},
			RunList: marshal.Slice[*Run]{
				{
					Command: marshal.Slice[*Command]{
						{
							Print: "exit 0",
							Exec:  "exit 0",
						},
					},
				},
			},
		}

		cachePath, err := task.taskInputCachePath(ctx)
		g.NoError(err)

		err = os.MkdirAll(filepath.Dir(cachePath), 0o700)
		g.NoError(err)

		err = os.WriteFile(cachePath, []byte("data writeonly"), 0o200)
		g.NoError(err)

		err = task.Execute(ctx)
		g.Should(be.ErrorEqual(
			err,
			fmt.Sprintf("checking cache: open %s: permission denied", cachePath),
		))

		runCount := strings.Count(buf.String(), "exit 0")
		if !g.Should(be.Equal(runCount, 0)) {
			t.Log(buf.String())
		}
	})
}

func TestTask_run_commands(t *testing.T) {
	g := ghost.New(t)

	var task Task

	runSuccess := &Run{
		Command: marshal.Slice[*Command]{{Exec: "exit 0"}},
	}

	err := task.run(Context{Logger: ui.Noop()}, runSuccess, stateRunning)
	g.NoError(err)

	runFailure := &Run{
		Command: marshal.Slice[*Command]{
			{Exec: "exit 0"},
			{Exec: "exit 1"},
		},
	}

	err = task.run(Context{Logger: ui.Noop()}, runFailure, stateRunning)
	g.Should(be.ErrorEqual(err, "exit status 1"))
}

func TestTask_run_sub_tasks(t *testing.T) {
	g := ghost.New(t)

	taskSuccess := Task{
		Name: "success",
		RunList: marshal.Slice[*Run]{
			&Run{Command: marshal.Slice[*Command]{{Exec: "exit 0"}}},
		},
	}

	taskFailure := Task{
		Name: "failure",
		RunList: marshal.Slice[*Run]{
			&Run{Command: marshal.Slice[*Command]{{Exec: "exit 1"}}},
		},
	}

	r := &Run{
		Tasks: []Task{taskSuccess},
	}

	task := Task{}

	err := task.run(Context{Logger: ui.Noop()}, r, stateRunning)
	g.NoError(err)

	r.Tasks = append(r.Tasks, taskFailure)

	err = task.run(Context{Logger: ui.Noop()}, r, stateRunning)
	g.Should(be.ErrorEqual(err, "exit status 1"))
}

func TestTask_run_environment(t *testing.T) {
	g := ghost.New(t)

	toBeUnset := "TO_BE_UNSET"
	toBeUnsetValue := "unsetvalue"

	toBeSet := "TO_BE_SET"
	toBeSetValue := "setvalue"

	t.Setenv(toBeSet, "")
	t.Setenv(toBeUnset, toBeUnsetValue)

	var task Task

	r := &Run{
		SetEnvironment: map[string]*string{
			toBeSet:   &toBeSetValue,
			toBeUnset: nil,
		},
	}

	err := task.run(Context{Logger: ui.Noop()}, r, stateRunning)
	g.NoError(err)

	g.Should(be.Equal(os.Getenv(toBeSet), toBeSetValue))

	got, ok := os.LookupEnv(toBeUnset)
	g.Should(be.Equal("", got))
	g.Check(!ok)
}

func TestTask_run_finally(t *testing.T) {
	g := ghost.New(t)

	task := Task{
		Finally: marshal.Slice[*Run]{
			&Run{Command: marshal.Slice[*Command]{{Exec: "exit 0"}}},
		},
	}

	var err error
	task.runFinally(Context{Logger: ui.Noop()}, &err)
	g.NoError(err)
}

func TestTask_run_finally_error(t *testing.T) {
	g := ghost.New(t)

	task := Task{
		Finally: marshal.Slice[*Run]{
			&Run{Command: marshal.Slice[*Command]{{Exec: "exit 1"}}},
		},
	}

	var err error
	task.runFinally(Context{Logger: ui.Noop()}, &err)
	g.Should(be.ErrorEqual(err, "exit status 1"))
}

func TestTask_run_finally_ui(t *testing.T) {
	g := ghost.New(t)

	taskName := "foo"
	command := "exit 0"

	want := new(bytes.Buffer)
	wantLogger := ui.New(ui.Config{
		Stderr:    want,
		Verbosity: ui.LevelVerbose,
	})

	wantLogger.PrintTaskFinally(taskName)
	wantLogger.PrintCommandWithParenthetical(command, "finally", taskName)

	got := new(bytes.Buffer)
	gotLogger := ui.New(ui.Config{
		Stderr:    got,
		Verbosity: ui.LevelVerbose,
	})

	task := Task{
		Name: taskName,
		Finally: marshal.Slice[*Run]{
			&Run{Command: marshal.Slice[*Command]{{Print: command}}},
		},
	}

	ctx := Context{
		Logger: gotLogger,
	}
	ctx = ctx.WithTask(&task)

	var err error
	task.runFinally(ctx, &err)
	g.NoError(err)

	g.Should(be.Equal(got.String(), want.String()))
}

func TestTask_run_finally_ui_error(t *testing.T) {
	g := ghost.New(t)

	taskName := "foo"
	command := "exit 1"
	wantErr := errors.New("exit status 1")

	want := new(bytes.Buffer)
	wantLogger := ui.New(ui.Config{
		Stderr:    want,
		Verbosity: ui.LevelVerbose,
	})

	wantLogger.PrintTaskFinally(taskName)
	wantLogger.PrintCommandWithParenthetical(command, "finally", taskName)
	wantLogger.PrintCommandError(wantErr)

	got := new(bytes.Buffer)
	gotLogger := ui.New(ui.Config{
		Stderr:    got,
		Verbosity: ui.LevelVerbose,
	})

	task := Task{
		Name: taskName,
		Finally: marshal.Slice[*Run]{
			&Run{Command: marshal.Slice[*Command]{{Exec: command, Print: command}}},
		},
	}

	ctx := Context{
		Logger: gotLogger,
	}
	ctx = ctx.WithTask(&task)

	var err error
	task.runFinally(ctx, &err)
	g.Should(be.ErrorEqual(err, "exit status 1"))

	g.Should(be.Equal(got.String(), want.String()))
}
