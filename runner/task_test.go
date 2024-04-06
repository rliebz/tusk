package runner

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

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

var executeTests = []struct {
	desc    string
	run     string
	finally string
	wantErr string
}{
	{
		"run error only",
		"exit 1",
		"exit 0",
		"exit status 1",
	},
	{
		"finally error only",
		"exit 0",
		"exit 1",
		"exit status 1",
	},
	{
		"run and finally error",
		"exit 1",
		"exit 2",
		"exit status 1",
	},
}

func TestTaskExecute_errors_returned(t *testing.T) {
	for _, tt := range executeTests {
		t.Run(tt.desc, func(t *testing.T) {
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
	ctx.PushTask(&task)

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
	ctx.PushTask(&task)

	var err error
	task.runFinally(ctx, &err)
	g.Should(be.ErrorEqual(err, "exit status 1"))

	g.Should(be.Equal(got.String(), want.String()))
}
