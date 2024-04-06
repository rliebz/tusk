package appcli

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	"github.com/rliebz/tusk/marshal"
	"github.com/rliebz/tusk/runner"
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
			g := ghost.New(t)

			originalArgs := os.Args
			t.Cleanup(func() {
				os.Args = originalArgs
			})

			// We only care about the "trailing" arg, second from last
			os.Args = []string{tt.trailing, "--"}

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

			var buf bytes.Buffer
			defaultComplete(&buf, c, app)

			g.Should(be.Equal(buf.String(), tt.want))
		})
	}
}

func TestCommandComplete(t *testing.T) {
	tests := []struct {
		name     string
		command  *cli.Command
		narg     int
		taskArgs runner.Args
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
			taskArgs: runner.Args{
				{
					Passable: runner.Passable{
						Name:          "first",
						ValuesAllowed: []string{"foo", "bar"},
					},
				},
				{
					Passable: runner.Passable{
						Name:          "second",
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
			taskArgs: runner.Args{
				{
					Passable: runner.Passable{
						Name:          "first",
						ValuesAllowed: []string{"foo", "bar"},
					},
				},
				{
					Passable: runner.Passable{
						Name:          "second",
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
			taskArgs: runner.Args{
				{
					Passable: runner.Passable{
						Name:          "foo",
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
			g := ghost.New(t)

			originalArgs := os.Args
			t.Cleanup(func() {
				os.Args = originalArgs
			})

			// We only care about the "trailing" arg, second from last
			os.Args = []string{tt.trailing, "--"}

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

			cfg := &runner.Config{
				Tasks: map[string]*runner.Task{
					cmd.Name: {
						Args: tt.taskArgs,
						Options: runner.Options{
							{
								Passable: runner.Passable{
									Name: "bool",
									Type: "bool",
								},
							},
							{
								Passable: runner.Passable{
									Name: "string",
								},
							},
							{
								Passable: runner.Passable{
									Name:          "values",
									ValuesAllowed: marshal.Slice[string]{"foo", "bar", "baz"},
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

			var buf bytes.Buffer
			commandComplete(&buf, c, cmd, cfg)

			g.Should(be.Equal(buf.String(), tt.want))
		})
	}
}

func TestPrintCommand(t *testing.T) {
	tests := []struct {
		name    string
		command *cli.Command
		want    string
	}{
		{
			name: "arg without usage",
			command: &cli.Command{
				Name: "my-cmd",
			},
			want: "my-cmd\n",
		},
		{
			name: "arg with usage",
			command: &cli.Command{
				Name:  "my-cmd",
				Usage: "My description",
			},
			want: "my-cmd:My description\n",
		},
		{
			name: "arg without usage escapes colon",
			command: &cli.Command{
				Name: "my:cmd",
			},
			want: "my\\:cmd\n",
		},
		{
			name: "arg with usage escapes colon",
			command: &cli.Command{
				Name:  "my:cmd",
				Usage: "My description",
			},
			want: "my\\:cmd:My description\n",
		},
		{
			name: "hidden",
			command: &cli.Command{
				Name:   "my-cmd",
				Usage:  "My description",
				Hidden: true,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			var buf bytes.Buffer
			printCommand(&buf, tt.command)

			g.Should(be.Equal(buf.String(), tt.want))
		})
	}
}

func TestPrintFlag(t *testing.T) {
	tests := []struct {
		name string
		flag cli.Flag
		want string
	}{
		{
			name: "flag without usage",
			flag: &cli.BoolFlag{
				Name: "my-flag",
			},
			want: "--my-flag\n",
		},
		{
			name: "arg with usage",
			flag: &cli.BoolFlag{
				Name:  "my-flag",
				Usage: "My description",
			},
			want: "--my-flag:My description\n",
		},
		{
			name: "arg without usage escapes colon",
			flag: &cli.BoolFlag{
				Name: "my:flag",
			},
			want: "--my\\:flag\n",
		},
		{
			name: "arg with usage escapes colon",
			flag: &cli.BoolFlag{
				Name:  "my:flag",
				Usage: "My description",
			},
			want: "--my\\:flag:My description\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			var c mockContext

			var buf bytes.Buffer
			printFlag(&buf, c, tt.flag)

			g.Should(be.Equal(buf.String(), tt.want))
		})
	}
}

func TestIsCompletingFlagArg(t *testing.T) {
	tests := []struct {
		flags []cli.Flag
		arg   string
		want  bool
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
		flag := "<no flags>"
		if len(tt.flags) >= 1 {
			flag = fmt.Sprintf("%T:%v", tt.flags[0], tt.flags[0].GetName())
		}
		if len(tt.flags) >= 2 {
			flag += fmt.Sprintf("|%T:%v", tt.flags[1], tt.flags[1].GetName())
		}

		t.Run(fmt.Sprintf("[%s] [%s]", flag, tt.arg), func(t *testing.T) {
			g := ghost.New(t)

			got := isCompletingFlagArg(tt.flags, tt.arg)
			g.Should(be.Equal(got, tt.want))
		})
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
