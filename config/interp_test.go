package config

import (
	"fmt"
	"testing"
)

var interpolatetests = []struct {
	testCase string
	cfgText  string
	passed   map[string]string
	taskName string
	expected string
}{
	{
		"happy path",
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
		`
options:
  foo:
    default: bar
tasks:
  mytask:
    run: echo bar
`,
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
		`
options:
  foo:
    default: bar
tasks:
  mytask:
    run: echo passed
`,
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
		`
options:
  foo:
    default: foovalue
  bar:
    default: barvalue
tasks:
  mytask:
    run: echo foovalue
  unused:
    run: echo ${bar}
`},

	{
		"no task specified",
		`
options:
  foo:
    default: bar
tasks:
  mytask:
    run: echo ${foo}
`,
		map[string]string{},
		"",
		`
options:
  foo:
    default: bar
tasks:
  mytask:
    run: echo ${foo}
`,
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
		`
options:
  foo:
    default: foovalue
  bar:
    default: foovalue
tasks:
  mytask:
    run: echo foovalue
`,
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
		`
options:
  foo:
    default: foovalue
tasks:
  mytask:
    options:
      bar:
        default: foovalue
    run: echo foovalue
`,
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
		`
options:
  foo:
    default: foovalue
tasks:
  mytask:
    options:
      foo:
        default: newvalue
    run: echo newvalue
`,
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
		`
tasks:
  unused:
    options:
      foo:
        default: foovalue
    run: echo barvalue
  mytask:
    options:
      foo:
        default: barvalue
    run: echo barvalue
`,
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
		`
options:
  foo:
    default: foovalue

tasks:
  pretask:
    run: echo foovalue
  mytask:
    run:
      task: pretask
`,
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
		`
options:
  foo:
    default: foovalue

tasks:
  roottask:
    run: echo foovalue
  pretask:
    run:
      task: roottask
  mytask:
    run:
      task: pretask
`,
	},

	{
		"nested sub-task dependencies with passed value",
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
		map[string]string{"foo": "passed"},
		"mytask",
		`
options:
  foo:
    default: foovalue

tasks:
  roottask:
    run: echo passed
  pretask:
    run:
      task: roottask
  mytask:
    run:
      task: pretask
`,
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
		`
tasks:
  roottask:
    options:
      foo:
        default: foovalue
    run: echo foovalue
  pretask:
    run:
      task: roottask
  mytask:
    run:
      task: pretask
`,
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
          value: barvalue
    run:
      when:
        equal:
          foo: true
      command: echo yo
`,
	},
}

func TestInterpolate(t *testing.T) {
	for _, tt := range interpolatetests {

		errString := fmt.Sprintf(
			"Interpolate(cfgText, passed, taskName) for %s:\n"+
				"cfgText: `%s`\npassed: %v\ntaskName: %s",
			tt.testCase, tt.cfgText, tt.passed, tt.taskName,
		)

		actualBytes, _, err := Interpolate([]byte(tt.cfgText), tt.passed, tt.taskName)
		if err != nil {
			t.Errorf("%s\nunexpected error: %s", errString, err)
			continue
		}

		actual := string(actualBytes)

		if tt.expected != actual {
			t.Errorf(
				"%s\nexpected: `%s`\nactual: `%s`\n",
				errString, tt.expected, actual,
			)
			continue
		}

	}
}

func TestInterpolate_no_redefining_sub_tasks(t *testing.T) {

	cfgText := `
tasks:
  one:
    options:
      foo:
        default: foovalue
  two:
    options:
      foo:
        default: barvalue
    run:
      task: one
  `

	if _, _, err := Interpolate([]byte(cfgText), nil, "foo"); err == nil {
		t.Errorf("Interpolate(cfgText, ...): expected error, got nil")
	}

}
