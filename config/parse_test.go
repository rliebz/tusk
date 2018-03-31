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
	args     []string
	flags    map[string]string
	taskName string
	expected task.RunList
}{
	{
		"argument interpolation",
		`
tasks:
  mytask:
    args:
      foo: {}
    run: echo ${foo}
`,
		[]string{"foovalue"},
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo foovalue"},
		}},
	},

	{
		"argument with global interpolation",
		`
options:
  foo:
    default: wrong
tasks:
  mytask:
    args:
      foo: {}
    run: echo ${foo}
`,
		[]string{"foovalue"},
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo foovalue"},
		}},
	},

	{
		"argument evaluated before option",
		`
tasks:
  mytask:
    args:
      foo: {}
    options:
      bar:
        default: ${foo}
    run: echo ${foo} ${bar}
`,
		[]string{"foovalue"},
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo foovalue foovalue"},
		}},
	},

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
		[]string{},
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
		[]string{},
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
		[]string{},
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
		[]string{},
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo ${bar}"},
		}},
	},

	{
		"multiple interpolation - global",
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
		[]string{},
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo foovalue"},
		}},
	},

	{
		"multiple interpolation - task specific",
		`
tasks:
  mytask:
    options:
      foo:
        default: foovalue
      bar:
        default: ${foo}
    run: echo ${bar}
`,
		[]string{},
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo foovalue"},
		}},
	},

	{
		"multiple interpolation - global + task specific",
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
		[]string{},
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
		[]string{},
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
		[]string{},
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
		[]string{},
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
		[]string{},
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
		[]string{},
		map[string]string{"foo": "passed"},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo passed-1-2"},
		}},
	},

	{
		"nested sub-task dependencies with args and options",
		`
options:
  foo:
    default: foovalue
tasks:
  roottask:
    args:
      one: {}
      two: {}
    options:
      foo:
        default: nope
    run:
      - echo ${foo}
      - echo ${one} ${two}
  pretask:
    args:
      one: {}
      two: {}
    options:
      foo:
        default: nope
    run:
      task:
        name: roottask
        args:
          - ${one}-2
          - ${two}-2
        options:
          foo: ${foo}-2
  mytask:
    run:
      task:
        name: pretask
        args:
          - onevalue
          - twovalue
        options:
          foo: ${foo}-1
`,
		[]string{},
		map[string]string{"foo": "passed"},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo passed-1-2"},
		}, {
			Command: marshal.StringList{"echo onevalue-2 twovalue-2"},
		}},
	},

	{
		"repeated sub-task call with different args",
		`
tasks:
  pretask:
    args:
      foo: {}
      bar: {}
    run: echo ${foo} ${bar}
  mytask:
    run:
      - task:
          name: pretask
          args:
            - one
            - two
      - task:
          name: pretask
          args:
            - three
            - four
`,
		[]string{},
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo one two"},
		}, {
			Command: marshal.StringList{"echo three four"},
		}},
	},

	{
		"repeated sub-task call with different options",
		`
tasks:
  pretask:
    options:
      foo: {}
    run: echo ${foo}
  mytask:
    run:
      - task:
          name: pretask
          options:
            foo: one
      - task:
          name: pretask
          options:
            foo: two
`,
		[]string{},
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo one"},
		}, {
			Command: marshal.StringList{"echo two"},
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
		[]string{},
		map[string]string{},
		"mytask",
		task.RunList{{
			Command: marshal.StringList{"echo foovalue"},
		}},
	},

	{
		"when clauses",
		`
tasks:
  mytask:
    run:
      when:
        - os:
            - os1
            - os2
        - command: echo hello
          os: os3
      command: echo goodbye
`,
		[]string{},
		map[string]string{},
		"mytask",
		task.RunList{{
			When: when.List{
				when.Create(when.WithOS("os1"), when.WithOS("os2")),
				when.Create(when.WithCommand("echo hello"), when.WithOS("os3")),
			},
			Command: marshal.StringList{"echo goodbye"},
		}},
	},

	{
		"when clause with dependencies",
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
		[]string{},
		map[string]string{},
		"mytask",
		task.RunList{{
			When: when.List{when.When{
				Equal: map[string]marshal.StringList{
					"foo": {"true"},
				},
			}},
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
		[]string{},
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
			tt.testCase, tt.taskName, tt.flags, tt.input,
		)

		cfg, err := ParseComplete([]byte(tt.input), tt.taskName, tt.args, tt.flags)
		if err != nil {
			t.Errorf(context+"unexpected error parsing text: %s", err)
			continue
		}

		actual := flattenRuns(cfg.Tasks[tt.taskName].RunList)

		if len(tt.expected) != len(actual) {
			t.Errorf(
				context+"task %q expected %d runs, actual: %d",
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
	t.Helper()

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

var invalidinterpolatetests = []struct {
	testCase string
	input    string
	args     []string
	flags    map[string]string
	taskName string
}{
	{
		"invalid yaml",
		`}{`,
		[]string{},
		map[string]string{},
		"mytask",
	},
	{
		"not passing required arg to subtask",
		`
tasks:
  one:
    args:
      foo: {}
    run: echo hello
  two:
    run:
      task:
        name: one
`,
		[]string{},
		map[string]string{},
		"two",
	},
	{
		"passing non-arg to subtask",
		`
tasks:
  one:
    run: echo hello
  two:
    run:
      task:
        name: one
        args: foo
`,
		[]string{},
		map[string]string{},
		"two",
	},
	{
		"not passing required option to subtask",
		`
tasks:
  one:
    options:
      foo: {required: true}
    run: echo hello
  two:
    run:
      task:
        name: one
`,
		[]string{},
		map[string]string{},
		"two",
	},
	{
		"passing non-option to subtask",
		`
tasks:
  one:
    run: echo hello
  two:
    run:
      task:
        name: one
        options: {wrong: foo}
`,
		[]string{},
		map[string]string{},
		"two",
	},
	{
		"passing global-option to subtask",
		`
options:
  foo:
    default: foovalue
tasks:
  one:
    run: echo ${foo}
  two:
    run:
      task:
        name: one
        options: {foo: replacement}
`,
		[]string{},
		map[string]string{},
		"two",
	},
	{
		"sub-task does not exist",
		`
tasks:
  mytask:
    run:
      task: fake
`,
		[]string{},
		map[string]string{},
		"mytask",
	},
	{
		"argument and option share name",
		`
tasks:
  mytask:
    args:
      foo: {}
    options:
      foo: {}
    run: echo oops
`,
		[]string{},
		map[string]string{"foo": "foovalue"},
		"mytask",
	},
	{
		"argument not passed",
		`
tasks:
  mytask:
    args:
      foo: {}
    run: echo oops
`,
		[]string{},
		map[string]string{},
		"mytask",
	},
	{
		"extra argument passed",
		`
tasks:
  mytask:
    run: echo oops
`,
		[]string{"foo"},
		map[string]string{},
		"mytask",
	},
}

func TestParseComplete_invalid(t *testing.T) {
	for _, tt := range invalidinterpolatetests {
		context := fmt.Sprintf(`
executing test case: %s
for task "%s" with parameters: %s
---
given input:
%s
---
`,
			tt.testCase, tt.taskName, tt.flags, tt.input,
		)

		_, err := ParseComplete([]byte(tt.input), tt.taskName, tt.args, tt.flags)
		if err == nil {
			t.Errorf(context+"expected error for test case: %s", tt.testCase)
			continue
		}
	}
}

func TestParseComplete_no_task(t *testing.T) {
	cfgText := []byte(`
options:
  foo:
    default: bar
  bar:
    default: ${foo}
tasks:
  mytask:
    run: echo ${bar}
`)

	cfg, err := ParseComplete(cfgText, "", []string{}, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error parsing text: %s", err)
	}

	expectedBar := "${foo}"
	actualBar := cfg.Options["bar"].DefaultValues[0].Value

	if expectedBar != actualBar {
		t.Errorf(
			`expected raw value for bar: "%s", actual: "%s"`,
			expectedBar, actualBar,
		)
	}

	expectedCommand := "echo ${bar}"
	actualCommand := cfg.Tasks["mytask"].RunList[0].Command[0]

	if expectedCommand != actualCommand {
		t.Errorf(
			`expected raw command for mytask: "%s", actual: "%s"`,
			expectedCommand, actualCommand,
		)
	}
}
