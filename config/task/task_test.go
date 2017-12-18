package task

import (
	"os"
	"testing"

	"github.com/rliebz/tusk/config/marshal"
	"github.com/rliebz/tusk/config/run"
	"github.com/rliebz/tusk/config/when"
	yaml "gopkg.in/yaml.v2"
)

func TestTask_UnmarshalYAML(t *testing.T) {
	y := []byte(`options: { one: {}, two: {} }`)
	task := Task{}

	if err := yaml.Unmarshal(y, &task); err != nil {
		t.Fatalf(
			`yaml.Unmarshal("%s", %+v): unexpected error: %s`,
			string(y), task, err,
		)
	}

	for _, expected := range []string{"one", "two"} {

		actual := task.Options[expected].Name
		if expected != actual {
			t.Errorf(
				`yaml.Unmarshal("%s", %+v): expected option name: %s, actual: %s`,
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

var shouldtests = []struct {
	desc     string
	input    *run.Run
	expected bool
}{
	{"nil when clause", &run.Run{When: nil}, true},
	{"empty when clause", &run.Run{When: &when.When{}}, true},
	{"true when clause", &run.Run{When: &when.When{
		Command: marshal.StringList{"test 1 = 1"},
	}}, true},
	{"false when clause", &run.Run{When: &when.When{
		Command: marshal.StringList{"test 1 = 0"},
	}}, false},
}

func TestTask_shouldRun(t *testing.T) {
	var task Task

	for _, tt := range shouldtests {
		actual, err := task.shouldRun(tt.input)
		if err != nil {
			t.Errorf(
				"task.shouldRun() for %s: unexpected error: %s",
				tt.desc, err,
			)
			continue
		}
		if tt.expected != actual {
			t.Errorf(
				"task.shouldRun() for %s: expected: %t, actual: %t",
				tt.desc, tt.expected, actual,
			)
		}
	}
}

func TestTask_run_commands(t *testing.T) {
	var task Task

	runSuccess := &run.Run{
		Command: marshal.StringList{"exit 0"},
	}

	if err := task.run(runSuccess); err != nil {
		t.Errorf(`task.run([exit 0]): unexpected error: %s`, err)
	}

	runFailure := &run.Run{
		Command: marshal.StringList{"exit 0", "exit 1"},
	}

	if err := task.run(runFailure); err == nil {
		t.Error(`task.run([exit 0, exit 1]): expected error, got nil`)
	}
}

func TestTask_run_sub_tasks(t *testing.T) {
	taskSuccess := &Task{
		Name: "success",
		Run: run.List{
			&run.Run{Command: marshal.StringList{"exit 0"}},
		},
	}

	taskFailure := &Task{
		Name: "failure",
		Run: run.List{
			&run.Run{Command: marshal.StringList{"exit 1"}},
		},
	}

	task := Task{
		SubTasks: []*Task{taskSuccess, taskFailure},
	}

	r := &run.Run{
		Task: run.SubTaskList{{Name: "success"}},
	}

	if err := task.run(r); err != nil {
		t.Errorf(`task.run([exit 0]): unexpected error: %s`, err)
	}

	r.Task = append(r.Task, &run.SubTask{Name: "failure"})

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

	r := &run.Run{
		Environment: map[string]*string{
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
