package config

import (
	"fmt"
	"testing"
)

var interpolatetests = []struct {
	cfgText  string
	passed   map[string]string
	taskName string
	expected string
}{
	// Happy path test case
	{
		`
options:
  foo:
    default: bar
tasks:
  mytask:
    run:
      - command: echo ${foo}
`,
		make(map[string]string),
		"mytask",
		`
options:
  foo:
    default: bar
tasks:
  mytask:
    run:
      - command: echo bar
`,
	},

	// One unused variable
	{
		`
options:
  foo:
    default: foovalue
  bar:
    default: barvalue
tasks:
  mytask:
    run:
      - command: echo ${foo}
  unused:
    run:
      - command: echo ${bar}
`,
		make(map[string]string),
		"mytask",
		`
options:
  foo:
    default: foovalue
  bar:
    default: barvalue
tasks:
  mytask:
    run:
      - command: echo foovalue
  unused:
    run:
      - command: echo ${bar}
`},

	// No task specified
	{
		`
options:
  foo:
    default: bar
tasks:
  mytask:
    run:
      - command: echo ${foo}
`,
		make(map[string]string),
		"",
		`
options:
  foo:
    default: bar
tasks:
  mytask:
    run:
      - command: echo ${foo}
`,
	},

	// Multiple interpolation - top level
	{
		`
options:
  foo:
    default: foovalue
  bar:
    default: ${foo}
tasks:
  mytask:
    run:
      - command: echo ${bar}
`,
		make(map[string]string),
		"mytask",
		`
options:
  foo:
    default: foovalue
  bar:
    default: foovalue
tasks:
  mytask:
    run:
      - command: echo foovalue
`,
	},

	// Multiple interpolation - task specific
	{
		`
options:
  foo:
    default: foovalue
tasks:
  mytask:
    options:
      bar:
        default: ${foo}
    run:
      - command: echo ${bar}
`,
		make(map[string]string),
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
    run:
      - command: echo foovalue
`,
	},

	// Sub-task dependencies
	{
		`
options:
  foo:
    default: foovalue

tasks:
  pretask:
    run:
      - command: echo ${foo}
  mytask:
    run:
      - task: pretask
`,
		map[string]string{"fast": "true"},
		"mytask",
		`
options:
  foo:
    default: foovalue

tasks:
  pretask:
    run:
      - command: echo foovalue
  mytask:
    run:
      - task: pretask
`,
	},

	// Nested sub-task dependencies
	{
		`
options:
  foo:
    default: foovalue

tasks:
  roottask:
    run:
      - command: echo ${foo}
  pretask:
    run:
      - task: roottask
  mytask:
    run:
      - task: pretask
`,
		map[string]string{"fast": "true"},
		"mytask",
		`
options:
  foo:
    default: foovalue

tasks:
  roottask:
    run:
      - command: echo foovalue
  pretask:
    run:
      - task: roottask
  mytask:
    run:
      - task: pretask
`,
	},
}

func TestInterpolate(t *testing.T) {
	for _, tt := range interpolatetests {

		errString := fmt.Sprintf(
			"Interpolate(cfgText, passed, taskName) failed.\n"+
				"cfgText: `%s`\npassed: %v\ntaskName: %s",
			tt.cfgText, tt.passed, tt.taskName,
		)

		actualBytes, err := Interpolate([]byte(tt.cfgText), tt.passed, tt.taskName)
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
