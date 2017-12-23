package config

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/rliebz/tusk/config/marshal"
	"github.com/rliebz/tusk/config/task"
	"github.com/rliebz/tusk/config/when"
)

var interpolatetests = []struct {
	testCase string
	input    string
	passed   map[string]string
	taskName string
	expected task.RunList
}{
	{
		"single task global interpolation",
		`
options:
  foo:
    default: bar
tasks:
  mytask:
    run: echo ${foo}
`,
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo bar"},
		}},
	},

	{
		"value passed",
		`
options:
  foo:
    default: bar
tasks:
  mytask:
    run: echo ${foo}
`,
		map[string]string{"foo": "passed"},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo passed"},
		}},
	},

	{
		"unused variable",
		`
options:
  foo:
    default: foovalue
  bar:
    default: barvalue
tasks:
  mytask:
    run: echo ${foo}
  unused:
    run: echo ${bar}
`,
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo foovalue"},
		}},
	},

	{
		"escaped interpolation over multiple iterations",
		`
options:
  foo:
    default: foovalue
  bar:
    default: ${foo}
tasks:
  pretask:
    run: echo $${bar}
  mytask:
    run:
      task: pretask
`,
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo ${bar}"},
		}},
	},

	{
		"multiple interpolation - top level",
		`
options:
  foo:
    default: foovalue
  bar:
    default: ${foo}
tasks:
  mytask:
    run: echo ${bar}
`,
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo foovalue"},
		}},
	},

	{
		"multiple interpolation - task specific",
		`
options:
  foo:
    default: foovalue
tasks:
  mytask:
    options:
      bar:
        default: ${foo}
    run: echo ${bar}
`,
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo foovalue"},
		}},
	},

	{
		"override global options with task-specific",
		`
options:
  foo:
    default: foovalue
tasks:
  mytask:
    options:
      foo:
        default: newvalue
    run: echo ${foo}
`,
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo newvalue"},
		}},
	},

	{
		"shared option defined per task",
		`
tasks:
  unused:
    options:
      foo:
        default: foovalue
    run: echo ${foo}
  mytask:
    options:
      foo:
        default: barvalue
    run: echo ${foo}
`,
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo barvalue"},
		}},
	},

	{
		"sub-task dependencies",
		`
options:
  foo:
    default: foovalue

tasks:
  pretask:
    run: echo ${foo}
  mytask:
    run:
      task: pretask
`,
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo foovalue"},
		}},
	},

	{
		"nested sub-task dependencies",
		`
options:
  foo:
    default: foovalue

tasks:
  roottask:
    run: echo ${foo}
  pretask:
    run:
      task: roottask
  mytask:
    run:
      task: pretask
`,
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo foovalue"},
		}},
	},

	{
		"nested sub-task dependencies with passed value",
		`
options:
  foo:
    default: foovalue

tasks:
  roottask:
    options:
      foo:
        default: nope
    run: echo ${foo}
  pretask:
    options:
      foo:
        default: nope
    run:
      task:
        name: roottask
        options:
          foo: ${foo}-2
  mytask:
    run:
      task:
        name: pretask
        options:
          foo: ${foo}-1
`,
		map[string]string{"foo": "passed"},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo passed-1-2"},
		}},
	},

	{
		"nested sub-task dependencies with sub-task-level options",
		`
tasks:
  roottask:
    options:
      foo:
        default: foovalue
    run: echo ${foo}
  pretask:
    run:
      task: roottask
  mytask:
    run:
      task: pretask
`,
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo foovalue"},
		}},
	},

	{
		"when dependencies",
		`
options:
  bar:
    default: barvalue

tasks:
  mytask:
    options:
      foo:
        default:
          when:
            equal:
              foo: true
          value: ${bar}
    run:
      when:
        equal:
          foo: true
      command: echo yo
`,
		map[string]string{},
		"mytask",
		task.RunList{{
			When: when.When{
				Equal: map[string]marshal.StringList{
					"foo": {"true"},
				},
			},
			Command: marshal.StringList{"echo yo"},
		}},
	},

	{
		"reference same global option in task/sub-task",
		`
options:
  foo:
    default: foovalue

tasks:
  one:
    run:
      - command: echo ${foo}
  two:
    run:
      - command: echo ${foo}
      - task: one
`,
		map[string]string{},
		"two",
		task.RunList{{
			Command: marshal.StringList{"echo foovalue"},
		}, {
			Command: marshal.StringList{"echo foovalue"},
		}},
	},
}

func TestParseComplete_interpolates(t *testing.T) {
	for _, tt := range interpolatetests {
		context := fmt.Sprintf(`
executing test case: %s
for task "%s" with parameters: %s
---
given input:
%s
---
`,
			tt.testCase, tt.taskName, tt.passed, tt.input,
		)

		cfg, err := ParseComplete([]byte(tt.input), tt.passed, tt.taskName)
		if err != nil {
			t.Errorf(context+"unexpected error parsing text: %s", err)
			continue
		}

		actual := flattenRuns(cfg.Tasks[tt.taskName].RunList)

		if len(tt.expected) != len(actual) {
			t.Errorf(
				context+`task "%s" expected %d tasks, actual: %d`,
				tt.taskName, len(tt.expected), len(actual),
			)
			return
		}

		for i := range tt.expected {
			runsAreEquivalent(t, context, tt.expected[i], actual[i])
		}
	}
}

func flattenRuns(runList task.RunList) task.RunList {
	var flattened task.RunList

	for _, run := range runList {
		if len(run.Tasks) == 0 {
			flattened = append(flattened, run)
			continue
		}

		for _, t := range run.Tasks {
			flattened = append(flattened, flattenRuns(t.RunList)...)
		}
	}

	return flattened
}

func runsAreEquivalent(t *testing.T, context string, r1 *task.Run, r2 *task.Run) {
	if !reflect.DeepEqual(r1.When, r2.When) {
		t.Errorf(
			context+"expected when: %#v\nactual: %#v",
			r1.When, r2.When,
		)
		return
	}

	if len(r1.Command) != len(r2.Command) {
		t.Errorf(
			context+`expected %d commands, actual: %d`,
			len(r1.Command), len(r2.Command),
		)
		return
	}

	for i := range r1.Command {
		if r1.Command[i] != r2.Command[i] {
			t.Errorf(
				context+"expected command: %s\nactual: %s",
				r1.Command[i], r2.Command[i],
			)
		}
	}
}
