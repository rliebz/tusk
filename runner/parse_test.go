package runner

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/tusk/marshal"
)

var interpolatetests = []struct {
	name     string
	input    string
	args     []string
	flags    map[string]string
	taskName string
	want     RunList
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
				Exec:  "echo passed-1-2",
				Print: "echo passed-1-2",
			}},
		}, {
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
				Exec:  "echo one two",
				Print: "echo one two",
			}},
		}, {
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
				Exec:  "echo one",
				Print: "echo one",
			}},
		}, {
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
				Exec:  "echo hello",
				Print: "echo hello",
			}},
		}, {
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
				Exec:  "echo first",
				Print: "echo first",
			}},
		}, {
			Command: CommandList{{
				Exec:  "echo pre-foovalue",
				Print: "echo pre-foovalue",
			}},
		}, {
			Command: CommandList{{
				Exec:  "echo pre-barvalue",
				Print: "echo pre-barvalue",
			}},
		}, {
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
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
		RunList{{
			When: WhenList{
				createWhen(withWhenOS("os1"), withWhenOS("os2")),
				createWhen(withWhenCommand("echo hello"), withWhenOS("os3")),
			},
			Command: CommandList{{
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
		RunList{{
			When: WhenList{When{
				Equal: map[string]marshal.StringList{
					"foo": {"true"},
				},
			}},
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
				Exec:  "echo foovalue",
				Print: "echo foovalue",
			}},
		}, {
			Command: CommandList{{
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
		RunList{{
			Command: CommandList{{
				Exec:  "echo foovalue",
				Print: "don't echo foovalue",
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
			g.NoErr(err)

			got := flattenRuns(cfg.Tasks[tt.taskName].AllRunItems())
			g.Should(ghost.DeepEqual(tt.want, got))
		})
	}
}

func flattenRuns(runList RunList) RunList {
	var flattened RunList

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
		t.Run(tt.name, func(t *testing.T) {
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
			if err == nil {
				t.Fatal("want error, got nil")
			}
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
	g.NoErr(err)

	wantBar := "${foo}"
	gotBar := cfg.Options[1].DefaultValues[0].Value
	g.Should(ghost.Equal(wantBar, gotBar))

	wantCommand := "echo ${bar}"
	gotCommand := cfg.Tasks["mytask"].RunList[0].Command[0].Exec
	g.Should(ghost.Equal(wantCommand, gotCommand))
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
	g.NoErr(err)

	g.Should(ghost.BeTrue(cfg.Tasks["quietCmd"].RunList[0].Command[0].Quiet))
	g.Should(ghost.BeTrue(cfg.Tasks["quietTask"].Quiet))
}
