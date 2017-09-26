package task

import (
	"testing"

	"github.com/rliebz/tusk/config/marshal"
	"github.com/rliebz/tusk/config/run"
	"github.com/rliebz/tusk/config/when"
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
