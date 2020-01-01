package appcli

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/urfave/cli"
)

func TestDefaultComplete(t *testing.T) {
	tests := []struct {
		name     string
		narg     int
		trailing string
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
				narg: tt.narg,
			}
			defaultComplete(&buf, c, app)

			if diff := cmp.Diff(tt.want, buf.String()); diff != "" {
				t.Errorf("completion output differs:\n%s", diff)
			}
		})
	}
}

func TestIsCompletingArg(t *testing.T) {
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
		actual := isCompletingArg(tt.flags, tt.arg)
		if tt.expected != actual {
			t.Errorf(
				"isCompletingArg(%#v, %s) => %t, want %t",
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
