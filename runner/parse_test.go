package runner

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"

	"github.com/rliebz/tusk/marshal"
)

var interpolatetests = []struct {
	name     string
	input    string
	args     []string
	flags    map[string]string
	taskName string
	want     marshal.Slice[*Run]
}{
	{
		"interpreter",
		`
interpreter: node

tasks:
  mytask:
    run: console.log('Hello')
`,
		[]string{},
		map[string]string{},
		"mytask",
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "console.log('Hello')",
				Print: "console.log('Hello')",
			}},
		}},
	},

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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue",
				Print: "echo foovalue",
			}},
		}},
	},

	{
		"multiple argument interpolation",
		`
tasks:
  mytask:
    args:
      foo: {}
      bar: {}
    run: echo ${foo} ${bar}
`,
		[]string{"foovalue", "barvalue"},
		map[string]string{},
		"mytask",
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue barvalue",
				Print: "echo foovalue barvalue",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue",
				Print: "echo foovalue",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue foovalue",
				Print: "echo foovalue foovalue",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo bar",
				Print: "echo bar",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo passed",
				Print: "echo passed",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue",
				Print: "echo foovalue",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo ${bar}",
				Print: "echo ${bar}",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue",
				Print: "echo foovalue",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue",
				Print: "echo foovalue",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue",
				Print: "echo foovalue",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo newvalue",
				Print: "echo newvalue",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo barvalue",
				Print: "echo barvalue",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue",
				Print: "echo foovalue",
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
		[]string{},
		map[string]string{},
		"mytask",
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue",
				Print: "echo foovalue",
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
		[]string{},
		map[string]string{"foo": "passed"},
		"mytask",
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo passed-1-2",
				Print: "echo passed-1-2",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo passed-1-2",
				Print: "echo passed-1-2",
			}},
		}, {
			Command: marshal.Slice[*Command]{{
				Exec:  "echo onevalue-2 twovalue-2",
				Print: "echo onevalue-2 twovalue-2",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo one two",
				Print: "echo one two",
			}},
		}, {
			Command: marshal.Slice[*Command]{{
				Exec:  "echo three four",
				Print: "echo three four",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo one",
				Print: "echo one",
			}},
		}, {
			Command: marshal.Slice[*Command]{{
				Exec:  "echo two",
				Print: "echo two",
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
		[]string{},
		map[string]string{},
		"mytask",
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue",
				Print: "echo foovalue",
			}},
		}},
	},

	{
		"finally dependencies",
		`
options:
  foo:
    default: foovalue
tasks:
  pretask:
    run: echo ${foo}
  mytask:
    run: echo hello
    finally:
      task: pretask
`,
		[]string{},
		map[string]string{},
		"mytask",
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo hello",
				Print: "echo hello",
			}},
		}, {
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue",
				Print: "echo foovalue",
			}},
		}},
	},

	{
		"sub-task finally dependencies",
		`
options:
  foo:
    default: foovalue
  bar:
    default: barvalue
tasks:
  pretask:
    run: echo pre-${foo}
    finally: echo pre-${bar}
  mytask:
    run: echo first
    finally:
      - task: pretask
      - command: echo done
`,
		[]string{},
		map[string]string{},
		"mytask",
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo first",
				Print: "echo first",
			}},
		}, {
			Command: marshal.Slice[*Command]{{
				Exec:  "echo pre-foovalue",
				Print: "echo pre-foovalue",
			}},
		}, {
			Command: marshal.Slice[*Command]{{
				Exec:  "echo pre-barvalue",
				Print: "echo pre-barvalue",
			}},
		}, {
			Command: marshal.Slice[*Command]{{
				Exec:  "echo done",
				Print: "echo done",
			}},
		}},
	},

	{
		"nested sub-task finally dependencies",
		`
tasks:
  roottask:
    options:
      foo:
        default: foovalue
    finally: echo ${foo}
  pretask:
    finally:
      task: roottask
  mytask:
    finally:
      task: pretask
`,
		[]string{},
		map[string]string{},
		"mytask",
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue",
				Print: "echo foovalue",
			}},
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
		marshal.Slice[*Run]{{
			When: WhenList{
				createWhen(withWhenOS("os1"), withWhenOS("os2")),
				createWhen(withWhenCommand("echo hello"), withWhenOS("os3")),
			},
			Command: marshal.Slice[*Command]{{
				Exec:  "echo goodbye",
				Print: "echo goodbye",
			}},
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
		marshal.Slice[*Run]{{
			When: WhenList{When{
				Equal: map[string]marshal.Slice[string]{
					"foo": {"true"},
				},
			}},
			Command: marshal.Slice[*Command]{{
				Exec:  "echo yo",
				Print: "echo yo",
			}},
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
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue",
				Print: "echo foovalue",
			}},
		}, {
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue",
				Print: "echo foovalue",
			}},
		}},
	},

	{
		"command with echo",
		`
tasks:
  mytask:
    args:
      foo: {}
    run:
      - command:
          exec: echo ${foo}
          print: don't echo ${foo}
`,
		[]string{"foovalue"},
		map[string]string{},
		"mytask",
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo foovalue",
				Print: "don't echo foovalue",
			}},
		}},
	},

	{
		"rewrite",
		`
tasks:
  mytask:
    options:
      foo:
        type: bool
        rewrite: newvalue
    run: echo ${foo}
`,
		[]string{},
		map[string]string{
			"foo": "true",
		},
		"mytask",
		marshal.Slice[*Run]{{
			Command: marshal.Slice[*Command]{{
				Exec:  "echo newvalue",
				Print: "echo newvalue",
			}},
		}},
	},
}

func TestParseComplete_interpolates(t *testing.T) {
	for _, tt := range interpolatetests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			t.Logf(`
executing test case: %s
for task %q with parameters: %s
---
given input:
%s
---
`,
				tt.name, tt.taskName, tt.flags, tt.input,
			)

			meta := &Metadata{
				CfgText: []byte(tt.input),
			}

			cfg, err := ParseComplete(meta, tt.taskName, tt.args, tt.flags)
			g.NoError(err)

			got := flattenRuns(cfg.Tasks[tt.taskName].AllRunItems())
			g.Should(be.DeepEqual(got, tt.want))
		})
	}
}

func flattenRuns(runList marshal.Slice[*Run]) marshal.Slice[*Run] {
	var flattened marshal.Slice[*Run]

	for _, run := range runList {
		if len(run.Tasks) == 0 {
			flattened = append(flattened, run)
			continue
		}

		for i := range run.Tasks {
			flattened = append(flattened, flattenRuns(run.Tasks[i].AllRunItems())...)
		}
	}

	return flattened
}

var invalidinterpolatetests = []struct {
	name     string
	input    string
	args     []string
	flags    map[string]string
	taskName string
	wantErr  string
}{
	{
		name:     "invalid yaml",
		input:    `}{`,
		taskName: "mytask",
		wantErr:  `yaml: did not find expected node content`,
	},
	{
		name: "not passing required arg to subtask",
		input: `
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
		taskName: "two",
		wantErr:  `subtask "one" requires 1 args but got 0`,
	},
	{
		name: "not passing correct arg type to subtask",
		input: `
tasks:
  one:
    args:
      foo:
        type: int
    run: echo hello
  two:
    run:
      task:
        name: one
        args: somevalue
`,
		taskName: "two",
		wantErr:  `value "somevalue" for argument "foo" is not of type "int"`,
	},
	{
		name: "passing non-arg to subtask",
		input: `
tasks:
  one:
    run: echo hello
  two:
    run:
      task:
        name: one
        args: foo
`,
		taskName: "two",
		wantErr:  `subtask "one" requires 0 args but got 1`,
	},
	{
		name: "not passing required option to subtask",
		input: `
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
		taskName: "two",
		wantErr:  `no value passed for required option: foo`,
	},
	{
		name: "not passing correct option type to subtask",
		input: `
tasks:
  one:
    options:
      foo: {type: float}
    run: echo hello
  two:
    run:
      task:
        name: one
        options: {foo: somevalue}
`,
		taskName: "two",
		wantErr:  `value "somevalue" for option "foo" is not of type "float"`,
	},
	{
		name: "passing non-option to subtask",
		input: `
tasks:
  one:
    run: echo hello
  two:
    run:
      task:
        name: one
        options: {wrong: foo}
`,
		taskName: "two",
		wantErr:  `option "wrong" cannot be passed to task "one"`,
	},
	{
		name: "passing global-option to subtask",
		input: `
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
		taskName: "two",
		wantErr:  `option "foo" cannot be passed to task "one"`,
	},
	{
		name: "sub-task does not exist",
		input: `
tasks:
  mytask:
    run:
      task: fake
`,
		taskName: "mytask",
		wantErr:  `sub-task "fake" does not exist`,
	},
	{
		name: "argument and option share name",
		input: `
tasks:
  mytask:
    args:
      foo: {}
    options:
      foo: {}
    run: echo oops
`,
		flags:    map[string]string{"foo": "foovalue"},
		taskName: "mytask",
		wantErr:  `argument and option "foo" must have unique names within a task`,
	},
	{
		name: "argument not passed",
		input: `
tasks:
  mytask:
    args:
      foo: {}
    run: echo oops
`,
		taskName: "mytask",
		wantErr:  `task "mytask" requires exactly 1 args, got 0`,
	},
	{
		name: "extra argument passed",
		input: `
tasks:
  mytask:
    run: echo oops
`,
		args:     []string{"foo"},
		taskName: "mytask",
		wantErr:  `task "mytask" requires exactly 0 args, got 1`,
	},

	{
		name: "non-boolean rewrite",
		input: `
tasks:
  mytask:
    options:
      foo:
        type: string
        rewrite: newvalue
    run: echo ${bar}
`,
		flags: map[string]string{
			"foo": "true",
		},
		taskName: "mytask",
		wantErr:  "rewrite may only be performed on boolean values",
	},

	{
		name: "chained rewrite",
		input: `
tasks:
  mytask:
    options:
      foo:
        type: bool
        rewrite: newvalue
      bar:
        default:
          when:
            equal: {foo: newvalue}
        rewrite: barvalue
    run: echo ${bar}
`,
		flags: map[string]string{
			"foo": "true",
		},
		taskName: "mytask",
		wantErr:  "rewrite may only be performed on boolean values",
	},
}

func TestParseComplete_invalid(t *testing.T) {
	for _, tt := range invalidinterpolatetests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			t.Logf(`
executing test case: %s
for task %q with parameters: %s
---
given input:
%s
---
`,
				tt.name, tt.taskName, tt.flags, tt.input,
			)

			meta := &Metadata{
				CfgText: []byte(tt.input),
			}

			_, err := ParseComplete(meta, tt.taskName, tt.args, tt.flags)
			g.Must(be.Error(err))
			g.Should(be.Equal(err.Error(), tt.wantErr))
		})
	}
}

func TestParseComplete_no_task(t *testing.T) {
	g := ghost.New(t)

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

	meta := &Metadata{
		CfgText: cfgText,
	}

	cfg, err := ParseComplete(meta, "", []string{}, map[string]string{})
	g.NoError(err)

	wantBar := "${foo}"
	gotBar := cfg.Options[1].DefaultValues[0].Value
	g.Should(be.Equal(gotBar, wantBar))

	wantCommand := "echo ${bar}"
	gotCommand := cfg.Tasks["mytask"].RunList[0].Command[0].Exec
	g.Should(be.Equal(gotCommand, wantCommand))
}

func TestParseComplete_quiet(t *testing.T) {
	g := ghost.New(t)

	cfgText := []byte(`
tasks:
  quietCmd:
    run:
      - exec: echo hello
        quiet: yes
  quietTask:
    quiet: yes
    run:
      - echo quiet
`)

	meta := &Metadata{
		CfgText: cfgText,
	}

	cfg, err := ParseComplete(meta, "", []string{}, map[string]string{})
	g.NoError(err)

	g.Should(be.True(cfg.Tasks["quietCmd"].RunList[0].Command[0].Quiet))
	g.Should(be.True(cfg.Tasks["quietTask"].Quiet))
}
