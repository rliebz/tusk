package task

import (
	"os"
	"testing"

	"github.com/rliebz/tusk/config/marshal"
	yaml "gopkg.in/yaml.v2"
)

func TestTask_UnmarshalYAML(t *testing.T) {
	y := []byte(`
options: { one: {}, two: {} }
args: { three: {}, four: {} }
`)
	task := Task{}

	if err := yaml.Unmarshal(y, &task); err != nil {
		t.Fatalf(
			`yaml.Unmarshal("%s", %+v): unexpected error: %s`,
			string(y), task, err,
		)
	}

	for _, expected := range []string{"one", "two"} {
		opt, ok := task.Options[expected]
		if !ok {
			t.Errorf(
				`yaml.Unmarshal(%q, %+v): did not find option %q`,
				string(y), task, expected,
			)
			continue
		}

		actual := opt.Name
		if expected != actual {
			t.Errorf(
				`yaml.Unmarshal("%s", %+v): expected option name: %s, actual: %s`,
				string(y), task, expected, actual,
			)
		}
	}

	for _, expected := range []string{"three", "four"} {
		arg, ok := task.Args[expected]
		if !ok {
			t.Errorf(
				`yaml.Unmarshal(%q, %+v): did not find arg %q`,
				string(y), task, expected,
			)
			continue
		}

		actual := arg.Name
		if expected != actual {
			t.Errorf(
				`yaml.Unmarshal(%q, %+v): expected arg name: %s, actual: %s`,
				string(y), task, expected, actual,
			)
		}
	}
}

func TestTask_UnmarshalYAML_invalid(t *testing.T) {
	y := []byte(`[invalid]`)
	task := Task{}

	if err := yaml.Unmarshal(y, &task); err == nil {
		t.Fatalf(
			"yaml.Unmarshal(%s, ...): expected error, actual nil", string(y),
		)
	}
}

func TestTask_UnmarshalYAML_option_and_arg_share_name(t *testing.T) {
	y := []byte(`
options: { foo: {} }
args: { foo: {} }
`)
	task := Task{}

	if err := yaml.Unmarshal(y, &task); err == nil {
		t.Fatalf(
			"yaml.Unmarshal(%s, ...): expected error, actual nil", string(y),
		)
	}
}

func TestTask_run_commands(t *testing.T) {
	var task Task

	runSuccess := &Run{
		Command: marshal.StringList{"exit 0"},
	}

	if err := task.run(runSuccess); err != nil {
		t.Errorf(`task.run([exit 0]): unexpected error: %s`, err)
	}

	runFailure := &Run{
		Command: marshal.StringList{"exit 0", "exit 1"},
	}

	if err := task.run(runFailure); err == nil {
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

	if err := task.run(r); err != nil {
		t.Errorf(`task.run([exit 0]): unexpected error: %s`, err)
	}

	r.Tasks = append(r.Tasks, taskFailure)

	if err := task.run(r); err == nil {
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

	if err := task.run(r); err != nil {
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
