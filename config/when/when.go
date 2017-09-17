package when

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/rliebz/tusk/config/marshal"
)

// When defines the conditions for running a task.
type When struct {
	Command marshal.StringList `yaml:",omitempty"`
	Exists  marshal.StringList `yaml:",omitempty"`
	OS      marshal.StringList `yaml:",omitempty"`

	Equal    map[string]marshal.StringList `yaml:",omitempty"`
	NotEqual map[string]marshal.StringList `yaml:"not_equal,omitempty"`
}

// Dependencies returns a list of options that are required explicitly.
// This does not include interpolations.
func (w *When) Dependencies() []string {

	var options []string

	if w == nil {
		return options
	}

	// Use a map to prevent duplicates
	references := make(map[string]struct{})

	for opt := range w.Equal {
		references[opt] = struct{}{}
	}
	for opt := range w.NotEqual {
		references[opt] = struct{}{}
	}

	for opt := range references {
		options = append(options, opt)
	}

	return options
}

// Validate returns an error if any when clauses fail.
func (w *When) Validate(vars map[string]string) error {
	if w == nil {
		return nil
	}

	for _, f := range w.Exists {
		// TODO: Should not exists errors be treated differently?
		if _, err := os.Stat(f); err != nil {
			return fmt.Errorf("file %s does not exist", f)
		}
	}

	if err := validateOS(runtime.GOOS, w.OS); err != nil {
		return err
	}

	for _, command := range w.Command {
		if err := testCommand(command); err != nil {
			return fmt.Errorf("test failed: %s", command)
		}
	}

	if err := validateEquality(vars, w.Equal, func(a, b string) bool {
		return a == b
	}); err != nil {
		return err
	}

	if err := validateEquality(vars, w.NotEqual, func(a, b string) bool {
		return a != b
	}); err != nil {
		return err
	}

	return nil
}

func validateOS(os string, required []string) error { // nolint: unparam
	// Nothing specified means any OS is fine
	if len(required) == 0 {
		return nil
	}

	// Otherwise, at least one must match
	for _, r := range required {
		if os == normalizeOS(r) {
			return nil
		}
	}

	return fmt.Errorf("current OS \"%s\" not listed in %v", os, required)
}

func normalizeOS(os string) string {
	lower := strings.ToLower(os)

	for _, alt := range []string{"mac", "macos", "osx"} {
		if lower == alt {
			return "darwin"
		}
	}

	for _, alt := range []string{"win"} {
		if lower == alt {
			return "windows"
		}
	}

	return lower
}

func testCommand(command string) error {
	_, err := exec.Command("sh", "-c", command).Output() // nolint: gas
	return err
}

func validateEquality(
	options map[string]string,
	cases map[string]marshal.StringList,
	compare func(string, string) bool,
) error {

	for name, values := range cases {
		for _, expected := range values {

			actual, ok := options[name]
			if !ok {
				return fmt.Errorf("option \"%s\" not defined", name)
			}

			if !compare(expected, actual) {
				return fmt.Errorf("option \"%s\" has value: %s", name, actual)
			}
		}
	}

	return nil
}
