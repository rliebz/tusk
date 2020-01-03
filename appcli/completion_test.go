package appcli

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rliebz/tusk/config"
	"github.com/rliebz/tusk/config/marshal"
	"github.com/rliebz/tusk/config/option"
	"github.com/rliebz/tusk/config/task"
	"github.com/urfave/cli"
)

func TestDefaultComplete(t *testing.T) {
	tests := []struct {
		name     string
		narg     int
		trailing string
		flagsSet []string
		want     string
	}{
		{
			name:     "invalid input",
			narg:     1,
			want:     ``,
			trailing: "foo",
		},
		{
			name: "default completion",
			want: `normal
foo:a foo command
--bool:a boolean value
--string:a string value
`,
			trailing: "tusk",
		},
		{
			name: "ignores set values",
			want: `normal
foo:a foo command
--string:a string value
`,
			trailing: "--bool",
			flagsSet: []string{"bool"},
		},
		{
			name: "flag completion",
			want: `file
`,
			trailing: "--string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func(args []string) {
				os.Args = args
			}(os.Args)
			// We only care about the "trailing" arg, second from last
			os.Args = []string{tt.trailing, "--"}

			var buf bytes.Buffer
			app := cli.NewApp()
			app.Commands = []cli.Command{
				{
					Name:  "foo",
					Usage: "a foo command",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "foo-flag",
							Usage: "a flag for foo",
						},
					},
				},
			}
			app.Flags = []cli.Flag{
				cli.BoolFlag{
					Name:  "bool",
					Usage: "a boolean value",
				},
				cli.StringFlag{
					Name:  "string",
					Usage: "a string value",
				},
			}

			c := mockContext{
				narg:  tt.narg,
				flags: tt.flagsSet,
			}
			defaultComplete(&buf, c, app)

			if diff := cmp.Diff(tt.want, buf.String()); diff != "" {
				t.Errorf("completion output differs:\n%s", diff)
			}
		})
	}
}

func TestCommandComplete(t *testing.T) {
	tests := []struct {
		name     string
		command  *cli.Command
		narg     int
		taskArgs option.Args
		flagsSet []string
		trailing string
		want     string
	}{
		{
			name: "default",
			want: `task-no-args
--bool:a boolean flag
--string:a string flag
--values:a flag with limited allowed values
`,
			trailing: "my-cmd",
		},
		{
			name: "first arg",
			want: `task-args
foo
bar
--bool:a boolean flag
--string:a string flag
--values:a flag with limited allowed values
`,
			taskArgs: option.Args{
				{
					Name: "first",
					ValueWithList: option.ValueWithList{
						ValuesAllowed: []string{"foo", "bar"},
					},
				},
				{
					Name: "second",
					ValueWithList: option.ValueWithList{
						ValuesAllowed: []string{"baz"},
					},
				},
			},
			trailing: "my-cmd",
		},
		{
			name: "second arg",
			want: `task-args
baz
--bool:a boolean flag
--string:a string flag
--values:a flag with limited allowed values
`,
			taskArgs: option.Args{
				{
					Name: "first",
					ValueWithList: option.ValueWithList{
						ValuesAllowed: []string{"foo", "bar"},
					},
				},
				{
					Name: "second",
					ValueWithList: option.ValueWithList{
						ValuesAllowed: []string{"baz"},
					},
				},
			},
			narg:     1,
			trailing: "my-cmd",
		},
		{
			name: "args with a flag set",
			want: `task-args
foo
bar
baz
--bool:a boolean flag
--values:a flag with limited allowed values
`,
			taskArgs: option.Args{
				{
					Name: "foo",
					ValueWithList: option.ValueWithList{
						ValuesAllowed: []string{"foo", "bar", "baz"},
					},
				},
			},
			flagsSet: []string{"string"},
			trailing: "my-cmd",
		},
		{
			name:     "string option",
			want:     "file\n",
			trailing: "--string",
		},
		{
			name: "string option with values",
			want: `value
foo
bar
baz
`,
			trailing: "--values",
		},
		{
			name: "boolean no values",
			want: `task-no-args
--string:a string flag
--values:a flag with limited allowed values
`,
			flagsSet: []string{"bool"},
			trailing: "--bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func(args []string) {
				os.Args = args
			}(os.Args)
			// We only care about the "trailing" arg, second from last
			os.Args = []string{tt.trailing, "--"}

			var buf bytes.Buffer

			cmd := &cli.Command{
				Name:  "my-cmd",
				Usage: "a command",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "bool",
						Usage: "a boolean flag",
					},
					cli.StringFlag{
						Name:  "string",
						Usage: "a string flag",
					},
					cli.StringFlag{
						Name:  "values",
						Usage: "a flag with limited allowed values",
					},
				},
			}

			cfg := &config.Config{
				Tasks: map[string]*task.Task{
					cmd.Name: {
						Args: tt.taskArgs,
						Options: option.Options{
							{Name: "bool", Type: "bool"},
							{Name: "string"},
							{
								Name: "values",
								ValueWithList: option.ValueWithList{
									ValuesAllowed: marshal.StringList{"foo", "bar", "baz"},
								},
							},
						},
					},
				},
			}

			c := mockContext{
				narg:  tt.narg,
				flags: tt.flagsSet,
			}

			commandComplete(&buf, c, cmd, cfg)

			if diff := cmp.Diff(tt.want, buf.String()); diff != "" {
				t.Errorf("completion output differs:\n%v", diff)
			}
		})
	}
}

func TestIsCompletingFlagArg(t *testing.T) {
	tests := []struct {
		flags    []cli.Flag
		arg      string
		expected bool
	}{
		{[]cli.Flag{}, "foo", false},
		{[]cli.Flag{}, "-f", false},
		{[]cli.Flag{}, "--foo", false},
		{[]cli.Flag{cli.BoolFlag{Name: "f, foo"}}, "-f", false},
		{[]cli.Flag{cli.BoolFlag{Name: "f, foo"}}, "--foo", false},
		{[]cli.Flag{cli.BoolTFlag{Name: "f, foo"}}, "-f", false},
		{[]cli.Flag{cli.BoolTFlag{Name: "f, foo"}}, "--foo", false},
		{[]cli.Flag{cli.StringFlag{Name: "f, foo"}}, "-f", true},
		{[]cli.Flag{cli.StringFlag{Name: "f, foo"}}, "--foo", true},
		{[]cli.Flag{cli.StringFlag{Name: "f, foo"}}, "--f", false},
		{[]cli.Flag{cli.StringFlag{Name: "b, bar"}}, "-f", false},
		{[]cli.Flag{cli.StringFlag{Name: "b, bar"}}, "--foo", false},
		{[]cli.Flag{cli.StringFlag{Name: "f, foo"}, cli.StringFlag{Name: "b, bar"}}, "-f", true},
		{[]cli.Flag{cli.StringFlag{Name: "f, foo"}, cli.StringFlag{Name: "b, bar"}}, "--foo", true},
	}

	for _, tt := range tests {
		actual := isCompletingFlagArg(tt.flags, tt.arg)
		if tt.expected != actual {
			t.Errorf(
				"isCompletingFlagArg(%#v, %s) => %t, want %t",
				tt.flags, tt.arg, actual, tt.expected,
			)
		}
	}
}

type mockContext struct {
	narg  int
	flags []string
}

func (m mockContext) NArg() int {
	return m.narg
}

func (m mockContext) IsSet(name string) bool {
	for _, flag := range m.flags {
		if flag == name {
			return true
		}
	}

	return false
}
