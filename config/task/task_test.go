package task

import (
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

func TestTask_runCommands(t *testing.T) {
	var task Task

	runSuccess := &run.Run{
		Command: marshal.StringList{"exit 0"},
	}

	if err := task.runCommands(runSuccess); err != nil {
		t.Errorf(
			`task.RunCommands([exit 0]): unexpected error: %s`, err,
		)
	}

	runFailure := &run.Run{
		Command: marshal.StringList{"exit 0", "exit 1"},
	}

	if err := task.runCommands(runFailure); err == nil {
		t.Error(
			`task.RunCommands([exit 0, exit 1]): expected error, got nil`,
		)
	}
}
