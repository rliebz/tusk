package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	"github.com/rliebz/tusk/ui"
	yaml "gopkg.in/yaml.v2"
)

func TestCommand_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want Command
	}{
		{
			"short-command",
			`example`,
			Command{
				Exec:  "example",
				Print: "example",
			},
		},
		{
			"do-no-echo",
			`exec: example`,
			Command{
				Exec:  "example",
				Print: "example",
			},
		},
		{
			"command-with-print",
			`{exec: something, print: echo example}`,
			Command{
				Exec:  "something",
				Print: "echo example",
			},
		},
		{
			"many-fields",
			`{exec: dovalue, print: printvalue, dir: dirvalue}`,
			Command{
				Exec:  "dovalue",
				Print: "printvalue",
				Dir:   "dirvalue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			var got Command
			err := yaml.UnmarshalStrict([]byte(tt.yaml), &got)
			g.NoError(err)

			g.Should(be.Equal(got, tt.want))
		})
	}
}

func TestCommand_exec(t *testing.T) {
	tests := []struct {
		name        string
		interpreter []string
		command     string
		want        []string
	}{
		{
			name:    "defaults",
			command: `echo "Hello world!"`,
			want:    []string{"sh", "-c", `echo "Hello world!"`},
		},
		{
			name:        "interpreter",
			interpreter: []string{"/usr/bin/env", "node", "-e"},
			command:     `console.log("Hello world!")`,
			want:        []string{"/usr/bin/env", "node", "-e", `console.log("Hello world!")`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			wd, err := os.Getwd()
			g.NoError(err)

			wantDir := filepath.Dir(wd)
			command := Command{
				Exec: tt.command,
				Dir:  "..",
			}

			t.Cleanup(func() { execCommand = exec.Command })
			execCommand = func(name string, arg ...string) *exec.Cmd {
				cs := []string{"-test.run=TestCommand_exec_helper", "--", name}
				cs = append(cs, arg...)
				cmd := exec.Command(os.Args[0], cs...)
				cmd.Env = []string{
					"TUSK_TEST_EXEC_COMMAND=1",
					"TUSK_TEST_COMMAND_ARGS=" + strings.Join(tt.want, ","),
					"TUSK_TEST_COMMAND_DIR=" + wantDir,
				}
				return cmd
			}

			ctx := Context{
				Logger:      ui.New(),
				Interpreter: tt.interpreter,
			}

			err = command.exec(ctx)
			g.NoError(err)
		})
	}
}

// TestCommand_exec_helper is a helper test that is called when mocking exec.
//
// The following environment variables can configure this function:
//
//   - TUSK_TEST_EXEC_COMMAND: Set to "1" to run this function.
//   - TUSK_TEST_COMMAND_ARGS: Set to a comma-separated list of expected command
//     arguments.
//   - TUSK_TEST_COMMAND_DIR: Set to the expected directory
func TestCommand_exec_helper(*testing.T) {
	if os.Getenv("TUSK_TEST_EXEC_COMMAND") != "1" {
		return
	}
	defer os.Exit(0)

	fail := func(msg interface{}) {
		fmt.Fprintln(os.Stdout, msg)
		os.Exit(1)
	}

	wantArgs := strings.Split(os.Getenv("TUSK_TEST_COMMAND_ARGS"), ",")
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	if diff := cmp.Diff(wantArgs, args); diff != "" {
		fail("arguments differ:\n" + diff)
	}

	dir, err := os.Getwd()
	if err != nil {
		fail("failed to get working dir: " + err.Error())
	}

	wantDir := os.Getenv("TUSK_TEST_COMMAND_DIR")
	if wantDir != "" && dir != wantDir {
		fail(fmt.Sprintf("want working dir %s, got %s", wantDir, dir))
	}
}

func TestCommandList_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want CommandList
	}{
		{
			"single-short-command",
			`example`,
			CommandList{
				{Exec: "example", Print: "example"},
			},
		},
		{
			"list-short-commands",
			`[one,two]`,
			CommandList{
				{Exec: "one", Print: "one"},
				{Exec: "two", Print: "two"},
			},
		},
		{
			"single-do-command",
			`exec: example`,
			CommandList{
				{Exec: "example", Print: "example"},
			},
		},
		{
			"list-do-commands",
			`[{exec: one},{exec: two}]`,
			CommandList{
				{Exec: "one", Print: "one"},
				{Exec: "two", Print: "two"},
			},
		},
		{
			"empty-list",
			`[]`,
			CommandList{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			var got CommandList
			err := yaml.UnmarshalStrict([]byte(tt.yaml), &got)
			g.NoError(err)

			g.Should(be.DeepEqual(got, tt.want))
		})
	}
}
