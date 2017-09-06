package task

import (
	"testing"

	"github.com/rliebz/tusk/appyaml"
)

var shouldtests = []struct {
	desc     string
	input    *run
	expected bool
}{
	{"nil when clause", &run{When: nil}, true},
	{"empty when clause", &run{When: &appyaml.When{}}, true},
	{"true when clause", &run{When: &appyaml.When{
		Command: appyaml.StringList{"test 1 = 1"},
	}}, true},
	{"false when clause", &run{When: &appyaml.When{
		Command: appyaml.StringList{"test 1 = 0"},
	}}, false},
}

func TestTask_shouldRun(t *testing.T) {

	var task Task

	for _, tt := range shouldtests {
		actual := task.shouldRun(tt.input)
		if tt.expected != actual {
			t.Errorf(
				"task.shouldRun() for %s: expected: %t, actual: %t",
				tt.desc, tt.expected, actual,
			)
		}
	}
}

var validatetests = []struct {
	desc      string
	input     *run
	shouldErr bool
}{
	{"neither command nor task values defined", &run{}, false},
	{"command values defined", &run{
		Command: appyaml.StringList{"foo"},
	}, false},
	{"task values defined", &run{
		Task: appyaml.StringList{"foo"},
	}, false},
	{"both command and task values defined", &run{
		Command: appyaml.StringList{"foo"},
		Task:    appyaml.StringList{"foo"},
	}, true},
}

func TestTask_validateRun(t *testing.T) {

	var task Task

	for _, tt := range validatetests {
		err := task.validateRun(tt.input)
		hasErr := err != nil
		if tt.shouldErr != hasErr {
			t.Errorf(
				"task.validateRun() for %s: expected err: %t, actual err: %s",
				tt.desc, tt.shouldErr, err,
			)
		}
	}
}
