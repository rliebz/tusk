package task

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
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
				Do:    "example",
				Print: "example",
			},
		},
		{
			"do-no-echo",
			`do: example`,
			Command{
				Do:    "example",
				Print: "example",
			},
		},
		{
			"command-with-print",
			`{do: something, print: echo example}`,
			Command{
				Do:    "something",
				Print: "echo example",
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
				{Do: "example", Print: "example"},
			},
		},
		{
			"list-short-commands",
			`[one,two]`,
			CommandList{
				{Do: "one", Print: "one"},
				{Do: "two", Print: "two"},
			},
		},
		{
			"single-do-command",
			`do: example`,
			CommandList{
				{Do: "example", Print: "example"},
			},
		},
		{
			"list-do-commands",
			`[{do: one},{do: two}]`,
			CommandList{
				{Do: "one", Print: "one"},
				{Do: "two", Print: "two"},
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

func TestExecCommand(t *testing.T) {
	command := "exit 0"

	if err := execCommand(command); err != nil {
		t.Fatalf(`execCommand("%s"): unexpected err: %s`, command, err)
	}
}

func TestExecCommand_error(t *testing.T) {
	command := "exit 1"
	errExpected := errors.New("exit status 1")
	if err := execCommand(command); err.Error() != errExpected.Error() {
		t.Fatalf(`execCommand("%s"): expected error "%s", actual "%s"`,
			command, errExpected, err,
		)
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
