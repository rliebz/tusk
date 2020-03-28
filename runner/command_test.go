package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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
			var got Command

			if err := yaml.UnmarshalStrict([]byte(tt.yaml), &got); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatched values:\n%s", diff)
			}
		})
	}
}

// TestCommand_exec_helper is a helper test that is called when mocking exec.
//
// The following environment variables can configure this function:
//
// - TUSK_WANT_TEST_COMMAND: Set to "1" to run this function.
// - TUSK_TEST_COMMAND_ARGS: Set to a comma-separated list of expected command
//   arguments.
// - TUSK_TEST_COMMAND_DIR: Set to the expected directory
func TestCommand_exec_helper(*testing.T) {
	if os.Getenv("TUSK_WANT_TEST_COMMAND") != "1" {
		return
	}
	defer os.Exit(0)

	fail := func(msg interface{}) {
		fmt.Fprintln(os.Stderr, msg)
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

func TestCommand_exec(t *testing.T) {
	wantCommand := "echo hello world"
	wantArgs := strings.Join([]string{getShell(), "-c", wantCommand}, ",")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	wantDir := filepath.Dir(wd)

	command := Command{
		Exec: wantCommand,
		Dir:  "..",
	}

	execCommand = func(name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=TestCommand_exec_helper", "--", name}
		cs = append(cs, arg...)
		cmd := exec.Command(os.Args[0], cs...) // nolint: gosec
		cmd.Env = []string{
			"TUSK_WANT_TEST_COMMAND=1",
			"TUSK_TEST_COMMAND_ARGS=" + wantArgs,
			"TUSK_TEST_COMMAND_DIR=" + wantDir,
		}
		return cmd
	}
	defer func() { execCommand = exec.Command }()

	if err := command.exec(Context{Logger: ui.Noop()}); err != nil {
		t.Fatal(err)
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
			var got CommandList

			if err := yaml.UnmarshalStrict([]byte(tt.yaml), &got); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatched values:\n%s", diff)
			}
		})
	}
}

func TestGetShell(t *testing.T) {
	originalShell := os.Getenv(shellEnvVar)
	defer func() {
		if err := os.Setenv(shellEnvVar, originalShell); err != nil {
			t.Errorf("Failed to reset SHELL environment variable: %v", err)
		}
	}()

	customShell := "/my/custom/sh"
	if err := os.Setenv(shellEnvVar, customShell); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}

	if actual := getShell(); actual != customShell {
		t.Errorf("getShell(): expected %v, actual %v", customShell, actual)
	}

	if err := os.Unsetenv(shellEnvVar); err != nil {
		t.Fatalf("Failed to unset environment variable: %v", err)
	}

	if actual := getShell(); actual != defaultShell {
		t.Errorf("getShell(): expected %v, actual %v", defaultShell, actual)
	}
}
