package task

import (
	"testing"

	"gitlab.com/rliebz/tusk/appyaml"
)

var shouldtests = []struct {
	desc     string
	input    *Run
	expected bool
}{
	{"nil when clause", &Run{When: nil}, true},
	{"empty when clause", &Run{When: &appyaml.When{}}, true},
	{"true when clause", &Run{When: &appyaml.When{
		Test: appyaml.StringList{Values: []string{"1 = 1"}},
	}}, true},
	{"false when clause", &Run{When: &appyaml.When{
		Test: appyaml.StringList{Values: []string{"1 = 0"}},
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
	input     *Run
	shouldErr bool
}{
	{"neither command nor task values defined", &Run{}, false},
	{"command values defined", &Run{
		Command: appyaml.StringList{Values: []string{"foo"}},
	}, false},
	{"task values defined", &Run{
		Task: appyaml.StringList{Values: []string{"foo"}},
	}, false},
	{"both command and task values defined", &Run{
		Command: appyaml.StringList{Values: []string{"foo"}},
		Task:    appyaml.StringList{Values: []string{"foo"}},
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
