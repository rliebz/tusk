package task

import (
	"bytes"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/rliebz/tusk/config/marshal"
	"github.com/rliebz/tusk/ui"
	yaml "gopkg.in/yaml.v2"
)

func TestTask_UnmarshalYAML(t *testing.T) {
	y := []byte(`
options: { one: {}, two: {} }
args: { three: {}, four: {} }
`)
	task := Task{}

	if err := yaml.UnmarshalStrict(y, &task); err != nil {
		t.Fatalf(
			`yaml.UnmarshalStrict("%s", %+v): unexpected error: %s`,
			string(y), task, err,
		)
	}

	for i, expected := range []string{"one", "two"} {
		actual := task.Options[i].Name
		if expected != actual {
			t.Errorf(
				`yaml.UnmarshalStrict("%s", %+v): expected option name: %s, actual: %s`,
				string(y), task, expected, actual,
			)
		}
	}

	for _, expected := range []string{"three", "four"} {
		arg, ok := task.Args.Lookup(expected)
		if !ok {
			t.Errorf(
				`yaml.UnmarshalStrict(%q, %+v): did not find arg %q`,
				string(y), task, expected,
			)
			continue
		}

		actual := arg.Name
		if expected != actual {
			t.Errorf(
				`yaml.UnmarshalStrict(%q, %+v): expected arg name: %s, actual: %s`,
				string(y), task, expected, actual,
			)
		}
	}
}

func TestTask_UnmarshalYAML_invalid(t *testing.T) {
	y := []byte(`[invalid]`)
	task := Task{}

	if err := yaml.UnmarshalStrict(y, &task); err == nil {
		t.Fatalf(
			"yaml.UnmarshalStrict(%s, ...): expected error, actual nil", string(y),
		)
	}
}

func TestTask_UnmarshalYAML_option_and_arg_share_name(t *testing.T) {
	y := []byte(`
options: { foo: {} }
args: { foo: {} }
`)
	task := Task{}

	if err := yaml.UnmarshalStrict(y, &task); err == nil {
		t.Fatalf(
			"yaml.UnmarshalStrict(%s, ...): expected error, actual nil", string(y),
		)
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
			run := Run{Command: marshal.StringList{tt.run}}
			finally := Run{Command: marshal.StringList{tt.finally}}
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
		Command: marshal.StringList{"exit 0"},
	}

	if err := task.run(RunContext{}, runSuccess, stateRunning); err != nil {
		t.Errorf(`task.run([exit 0]): unexpected error: %s`, err)
	}

	runFailure := &Run{
		Command: marshal.StringList{"exit 0", "exit 1"},
	}

	if err := task.run(RunContext{}, runFailure, stateRunning); err == nil {
		t.Error(`task.run([exit 0, exit 1]): expected error, got nil`)
	}
}

func TestTask_run_sub_tasks(t *testing.T) {
	taskSuccess := Task{
		Name: "success",
		RunList: RunList{
			&Run{Command: marshal.StringList{"exit 0"}},
		},
	}

	taskFailure := Task{
		Name: "failure",
		RunList: RunList{
			&Run{Command: marshal.StringList{"exit 1"}},
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
			&Run{Command: marshal.StringList{"exit 0"}},
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
			&Run{Command: marshal.StringList{"exit 1"}},
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
			&Run{Command: marshal.StringList{command}},
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
			&Run{Command: marshal.StringList{command}},
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
