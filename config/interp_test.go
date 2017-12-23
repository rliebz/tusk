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
			Tasks: []task.Task{{
				RunList: task.RunList{{
					Command: marshal.StringList{"echo ${bar}"},
				}},
			}},
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
			Tasks: []task.Task{{
				RunList: task.RunList{{
					Command: marshal.StringList{"echo foovalue"},
				}},
			}},
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
			Tasks: []task.Task{{
				RunList: task.RunList{{
					Tasks: []task.Task{{
						RunList: task.RunList{{
							Command: marshal.StringList{"echo foovalue"},
						}},
					}},
				}},
			}},
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
			Tasks: []task.Task{{
				RunList: task.RunList{{
					Tasks: []task.Task{{
						RunList: task.RunList{{
							Command: marshal.StringList{"echo passed-1-2"},
						}},
					}},
				}},
			}},
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
			Tasks: []task.Task{{
				RunList: task.RunList{{
					Tasks: []task.Task{{
						RunList: task.RunList{{
							Command: marshal.StringList{"echo foovalue"},
						}},
					}},
				}},
			}},
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
			Tasks: []task.Task{{
				RunList: task.RunList{{
					Command: marshal.StringList{"echo foovalue"},
				}},
			}},
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

		actual, err := ParseComplete([]byte(tt.input), tt.passed, tt.taskName)
		if err != nil {
			t.Errorf(context+"unexpected error parsing text: %s", err)
			continue
		}

		expectedTask := &task.Task{
			Name:    tt.taskName,
			RunList: tt.expected,
		}

		tasksRunEquivalently(t, context, expectedTask, actual.Tasks[tt.taskName])
	}
}

func tasksRunEquivalently(t *testing.T, context string, t1 *task.Task, t2 *task.Task) {
	if len(t1.RunList) != len(t2.RunList) {
		t.Errorf(
			context+`task "%s" expected %d tasks, actual: %d`,
			t2.Name, len(t1.RunList), len(t2.RunList),
		)
		return
	}

	for i := range t1.RunList {
		runsAreEquivalent(t, context, t1.RunList[i], t2.RunList[i])
	}

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

	if len(r1.Tasks) != len(r2.Tasks) {
		t.Errorf(
			context+`expected %d subtasks, actual: %d`,
			len(r1.Tasks), len(r2.Tasks),
		)
		return
	}

	for i := range r1.Tasks {
		tasksRunEquivalently(t, context, &r1.Tasks[i], &r2.Tasks[i])
	}
}
