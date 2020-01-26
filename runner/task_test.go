package runner

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rliebz/tusk/ui"
	yaml "gopkg.in/yaml.v2"
	"gotest.tools/v3/assert"
)

func TestTask_UnmarshalYAML(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

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
					{Name: "one"},
					{Name: "two"},
				},
				Args: Args{
					{Name: "three"},
					{Name: "four"},
				},
			},
		},
		{
			name:  "include",
			input: fmt.Sprintf(`{include: %q}`, testdata("included.yml")),
			want: Task{
				Usage: "A valid example of an included task",
				RunList: RunList{{Command: CommandList{{
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
			var got Task
			err := yaml.UnmarshalStrict([]byte(tt.input), &got)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NilError(t, err)
			}

			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(Option{})); diff != "" {
				t.Errorf("parsed task differs from expected:\n%s", diff)
			}
		})
	}
}

var executeTests = []struct {
	desc     string
	run      string
	finally  string
	expected error
}{
	{
		"run error only",
		"exit 1",
		"exit 0",
		errors.New("exit status 1"),
	},
	{
		"finally error only",
		"exit 0",
		"exit 1",
		errors.New("exit status 1"),
	},
	{
		"run and finally error",
		"exit 1",
		"exit 2",
		errors.New("exit status 1"),
	},
}

func TestTaskExecute_errors_returned(t *testing.T) {
	for _, tt := range executeTests {
		t.Run(tt.desc, func(t *testing.T) {
			run := Run{Command: CommandList{{Exec: tt.run}}}
			finally := Run{Command: CommandList{{Exec: tt.finally}}}
			task := Task{
				RunList: RunList{&run},
				Finally: RunList{&finally},
			}

			actual := task.Execute(RunContext{})
			if actual.Error() != tt.expected.Error() {
				t.Errorf("want error %s, got %s", tt.expected, actual)
			}
		})
	}
}

func TestTask_run_commands(t *testing.T) {
	var task Task

	runSuccess := &Run{
		Command: CommandList{{Exec: "exit 0"}},
	}

	if err := task.run(RunContext{}, runSuccess, stateRunning); err != nil {
		t.Errorf(`task.run([exit 0]): unexpected error: %s`, err)
	}

	runFailure := &Run{
		Command: CommandList{
			{Exec: "exit 0"},
			{Exec: "exit 1"},
		},
	}

	if err := task.run(RunContext{}, runFailure, stateRunning); err == nil {
		t.Error(`task.run([exit 0, exit 1]): expected error, got nil`)
	}
}

func TestTask_run_sub_tasks(t *testing.T) {
	taskSuccess := Task{
		Name: "success",
		RunList: RunList{
			&Run{Command: CommandList{{Exec: "exit 0"}}},
		},
	}

	taskFailure := Task{
		Name: "failure",
		RunList: RunList{
			&Run{Command: CommandList{{Exec: "exit 1"}}},
		},
	}

	r := &Run{
		Tasks: []Task{taskSuccess},
	}

	task := Task{}

	if err := task.run(RunContext{}, r, stateRunning); err != nil {
		t.Errorf(`task.run([exit 0]): unexpected error: %s`, err)
	}

	r.Tasks = append(r.Tasks, taskFailure)

	if err := task.run(RunContext{}, r, stateRunning); err == nil {
		t.Error(`task.run([exit 0, exit 1]): expected error, got nil`)
	}
}

func TestTask_run_environment(t *testing.T) {
	toBeUnset := "TO_BE_UNSET"
	toBeUnsetValue := "unsetvalue"

	toBeSet := "TO_BE_SET"
	toBeSetValue := "setvalue"

	if err := os.Setenv(toBeUnset, toBeUnsetValue); err != nil {
		t.Fatalf(
			"os.Setenv(%s, %s): unexpected error: %v",
			toBeUnset, toBeUnsetValue, err,
		)
	}

	defer func() {
		if err := os.Unsetenv(toBeSet); err != nil {
			t.Errorf(
				"os.Unsetenv(%s): unexpected error: %v",
				toBeSet, err,
			)
		}
		if err := os.Unsetenv(toBeUnset); err != nil {
			t.Errorf(
				"os.Unsetenv(%s): unexpected error: %v",
				toBeUnset, err,
			)
		}
	}()

	var task Task

	r := &Run{
		SetEnvironment: map[string]*string{
			toBeSet:   &toBeSetValue,
			toBeUnset: nil,
		},
	}

	if err := task.run(RunContext{}, r, stateRunning); err != nil {
		t.Errorf("task.run(): unexpected error: %s", err)
	}

	if actual := os.Getenv(toBeSet); toBeSetValue != actual {
		t.Errorf(
			`value for %s: expected: "%s", actual: "%s"`,
			toBeSet, toBeSetValue, actual,
		)
	}

	if actual, isSet := os.LookupEnv(toBeUnset); isSet {
		t.Errorf(
			`value for %s: expected env var to be unset, actual: %s`,
			toBeUnset, actual,
		)
	}
}

func TestTask_run_finally(t *testing.T) {
	task := Task{
		Finally: RunList{
			&Run{Command: CommandList{{Exec: "exit 0"}}},
		},
	}

	var err error
	if task.runFinally(RunContext{}, &err); err != nil {
		t.Errorf("task.runFinally(): unexpected error: %s", err)
	}
}

func TestTask_run_finally_error(t *testing.T) {
	task := Task{
		Finally: RunList{
			&Run{Command: CommandList{{Exec: "exit 1"}}},
		},
	}

	var err error
	if task.runFinally(RunContext{}, &err); err == nil {
		t.Error("task.runFinally(): want error for exit status 1, got nil")
	}
}

func TestTask_run_finally_ui(t *testing.T) {
	defer func(level ui.VerbosityLevel) {
		ui.LoggerStderr.SetOutput(os.Stderr)
		ui.Verbosity = level
	}(ui.Verbosity)

	ui.LoggerStderr = log.New(os.Stderr, "", 0)
	ui.Verbosity = ui.VerbosityLevelVerbose
	taskName := "foo"
	command := "exit 0"

	bufExpected := new(bytes.Buffer)
	ui.LoggerStderr.SetOutput(bufExpected)
	ui.PrintTaskFinally(taskName)
	ui.PrintCommandWithParenthetical(command, "finally", taskName)
	expected := bufExpected.String()

	bufActual := new(bytes.Buffer)
	ui.LoggerStderr.SetOutput(bufActual)

	task := Task{
		Name: taskName,
		Finally: RunList{
			&Run{Command: CommandList{{Print: command}}},
		},
	}

	ctx := RunContext{}
	ctx.PushTask(&task)

	var err error
	if task.runFinally(ctx, &err); err != nil {
		t.Fatalf("task.runFinally(): unexpected error: %s", err)
	}

	actual := bufActual.String()

	if expected != actual {
		t.Fatalf(
			"task.runFinally(): want to print %q, got %q",
			expected, actual,
		)
	}
}

func TestTask_run_finally_ui_fails(t *testing.T) {
	defer func(l *log.Logger, ll ui.VerbosityLevel) {
		ui.LoggerStderr = l
		ui.Verbosity = ll
	}(ui.LoggerStderr, ui.Verbosity)

	ui.LoggerStderr = log.New(os.Stderr, "", 0)
	ui.Verbosity = ui.VerbosityLevelVerbose
	taskName := "foo"
	command := "exit 1"
	errExpected := errors.New("exit status 1")

	bufExpected := new(bytes.Buffer)
	ui.LoggerStderr.SetOutput(bufExpected)
	ui.PrintTaskFinally(taskName)
	ui.PrintCommandWithParenthetical(command, "finally", taskName)
	ui.PrintCommandError(errExpected)
	expected := bufExpected.String()

	bufActual := new(bytes.Buffer)
	ui.LoggerStderr.SetOutput(bufActual)

	task := Task{
		Name: taskName,
		Finally: RunList{
			&Run{Command: CommandList{{Exec: command, Print: command}}},
		},
	}

	ctx := RunContext{}
	ctx.PushTask(&task)

	var err error
	if task.runFinally(ctx, &err); err == nil {
		t.Error("task.runFinally(): want error for exit status 1, got nil")
	}

	actual := bufActual.String()

	if expected != actual {
		t.Fatalf(
			"task.runFinally(): want to print %q, got %q",
			expected, actual,
		)
	}
}
