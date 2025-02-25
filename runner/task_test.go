package runner

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	yaml "gopkg.in/yaml.v2"

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

func TestTaskExecute_skip(t *testing.T) {
	g := ghost.New(t)
	dir := useTempDir(t)

	createFile := func(name string, ago time.Duration) {
		path := filepath.Join(dir, name)
		err := os.WriteFile(path, []byte(name), 0o600)
		g.NoError(err)

		err = os.Chtimes(path, time.Time{}, time.Now().Add(-ago))
		g.NoError(err)
	}

	createFile("old.txt", 6*time.Hour)
	createFile("new.txt", 3*time.Hour)

	tests := []struct {
		name    string
		run     string
		source  []string
		target  []string
		wantErr string
	}{
		{
			name:   "with skip",
			run:    "exit 1",
			source: []string{"old.txt"},
			target: []string{"new.txt"},
		},
		{
			name:    "without skip",
			run:     "exit 1",
			source:  []string{"new.txt"},
			target:  []string{"old.txt"},
			wantErr: "exit status 1",
		},
		{
			name:    "with missing source glob",
			run:     "exit 1",
			source:  []string{"fake.txt"},
			target:  []string{"old.txt"},
			wantErr: "no source files found matching pattern: fake.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			run := Run{Command: marshal.Slice[*Command]{{Exec: tt.run}}}
			task := Task{
				RunList: marshal.Slice[*Run]{&run},
				Source:  tt.source,
				Target:  tt.target,
			}

			err := task.Execute(Context{Logger: ui.Noop()})
			if tt.wantErr != "" {
				g.Should(be.ErrorEqual(err, tt.wantErr))
				return
			}
			g.NoError(err)
		})
	}
}

func TestTaskExecute_errors_returned(t *testing.T) {
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

func TestTask_isUpToDate(t *testing.T) {
	g := ghost.New(t)

	dir := useTempDir(t)

	createFile := func(name string, ago time.Duration) {
		path := filepath.Join(dir, name)
		err := os.WriteFile(path, []byte(name), 0o600)
		g.NoError(err)

		err = os.Chtimes(path, time.Time{}, time.Now().Add(-ago))
		g.NoError(err)
	}

	createFile("a1.txt", 6*time.Hour)
	createFile("a2.txt", 5*time.Hour)
	createFile("a3.txt", 4*time.Hour)
	createFile("b1.txt", 3*time.Hour)
	createFile("b2.txt", 2*time.Hour)
	createFile("b3.txt", 1*time.Hour)

	tests := []struct {
		name         string
		source       []string
		target       []string
		wantUpToDate bool
		wantError    string
	}{
		{
			name:         "no source or target",
			wantUpToDate: false,
		},
		{
			name:         "source equal target",
			source:       []string{"a1.txt"},
			target:       []string{"a1.txt"},
			wantUpToDate: true,
		},
		{
			name:         "source before target",
			source:       []string{"a1.txt"},
			target:       []string{"a2.txt"},
			wantUpToDate: true,
		},
		{
			name:         "source after target",
			source:       []string{"a2.txt"},
			target:       []string{"a1.txt"},
			wantUpToDate: false,
		},

		{
			name:         "source between targets",
			source:       []string{"a2.txt"},
			target:       []string{"a1.txt", "a3.txt"},
			wantUpToDate: false,
		},
		{
			name:         "target between sources",
			source:       []string{"a1.txt", "a3.txt"},
			target:       []string{"a2.txt"},
			wantUpToDate: false,
		},
		{
			name:         "source glob before target",
			source:       []string{"a*.txt"},
			target:       []string{"b1.txt"},
			wantUpToDate: true,
		},
		{
			name:         "source glob after target",
			source:       []string{"b*.txt"},
			target:       []string{"a1.txt"},
			wantUpToDate: false,
		},
		{
			name:         "target does not exist",
			source:       []string{"a1.txt"},
			target:       []string{"fake.txt"},
			wantUpToDate: false,
		},
		{
			name:      "source does not exist",
			source:    []string{"fake.txt"},
			target:    []string{"a1.txt"},
			wantError: "no source files found matching pattern: fake.txt",
		},
		{
			name:      "invalid source pattern",
			source:    []string{`\`},
			target:    []string{"a1.txt"},
			wantError: `syntax error in source pattern: `,
		},
		{
			name:      "invalid target pattern",
			source:    []string{"a1.txt"},
			target:    []string{`\`},
			wantError: `syntax error in target pattern: `,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			task := Task{
				Source: tt.source,
				Target: tt.target,
			}

			got, err := task.isUpToDate(Context{
				Logger: ui.Noop(),
			})
			if tt.wantError != "" {
				g.Should(be.ErrorContaining(err, tt.wantError))
				return
			}
			g.NoError(err)

			g.Should(be.Equal(got, tt.wantUpToDate))
		})
	}
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
	g.Should(be.False(ok))
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
	logger := ui.New()
	logger.Verbosity = ui.VerbosityLevelVerbose
	logger.Stderr = want

	logger.PrintTaskFinally(taskName)
	logger.PrintCommandWithParenthetical(command, "finally", taskName)

	got := new(bytes.Buffer)
	logger.Stderr = got

	task := Task{
		Name: taskName,
		Finally: marshal.Slice[*Run]{
			&Run{Command: marshal.Slice[*Command]{{Print: command}}},
		},
	}

	ctx := Context{
		Logger: logger,
	}
	ctx = ctx.WithTask(&task)

	var err error
	task.runFinally(ctx, &err)
	g.NoError(err)

	g.Should(be.Equal(got.String(), want.String()))
}

func TestTask_run_finally_ui_fails(t *testing.T) {
	g := ghost.New(t)

	taskName := "foo"
	command := "exit 1"
	wantErr := errors.New("exit status 1")

	want := new(bytes.Buffer)
	logger := ui.New()
	logger.Verbosity = ui.VerbosityLevelVerbose
	logger.Stderr = want

	logger.PrintTaskFinally(taskName)
	logger.PrintCommandWithParenthetical(command, "finally", taskName)
	logger.PrintCommandError(wantErr)

	got := new(bytes.Buffer)
	logger.Stderr = got

	task := Task{
		Name: taskName,
		Finally: marshal.Slice[*Run]{
			&Run{Command: marshal.Slice[*Command]{{Exec: command, Print: command}}},
		},
	}

	ctx := Context{
		Logger: logger,
	}
	ctx = ctx.WithTask(&task)

	var err error
	task.runFinally(ctx, &err)
	g.Should(be.ErrorEqual(err, "exit status 1"))

	g.Should(be.Equal(got.String(), want.String()))
}
