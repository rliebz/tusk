package task

import (
	"testing"

	"github.com/rliebz/tusk/config/task/marshal"
	"github.com/rliebz/tusk/config/task/run"
	"github.com/rliebz/tusk/config/task/when"
)

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
	input     *run.Run
	shouldErr bool
}{
	{"neither command nor task values defined", &run.Run{}, false},
	{"command values defined", &run.Run{
		Command: marshal.StringList{"foo"},
	}, false},
	{"task values defined", &run.Run{
		Task: marshal.StringList{"foo"},
	}, false},
	{"both command and task values defined", &run.Run{
		Command: marshal.StringList{"foo"},
		Task:    marshal.StringList{"foo"},
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
