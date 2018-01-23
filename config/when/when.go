package when

import (
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
	if w == nil {
		return nil
	}

	// Use a map to prevent duplicates
	references := make(map[string]struct{})

	for opt := range w.Equal {
		references[opt] = struct{}{}
	}
	for opt := range w.NotEqual {
		references[opt] = struct{}{}
	}

	options := make([]string, 0, len(references))
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
		if _, err := os.Stat(f); err != nil {
			if os.IsNotExist(err) {
				return newCondFailErrorf(`file "%s" does not exist`, f)
			}
			return err
		}
	}

	if err := validateOS(runtime.GOOS, w.OS); err != nil {
		return err
	}

	for _, command := range w.Command {
		if err := testCommand(command); err != nil {
			return newCondFailErrorf(`test failed: %s`, command)
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

	return newCondFailErrorf(`current OS "%s" not listed in %v`, os, required)
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
				return newCondFailErrorf(`option "%s" not defined`, name)
			}

			if !compare(expected, actual) {
				return newCondFailErrorf(
					`option "%s" expected value "%s", but received "%s"`,
					name, expected, actual,
				)
			}
		}
	}

	return nil
}

// List is a list of when items with custom yaml unmarshalling.
type List []When

// UnmarshalYAML allows single items to be used as lists.
func (l *List) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var whenSlice []When
	sliceCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&whenSlice) },
		Assign:    func() { *l = whenSlice },
	}

	var whenItem When
	itemCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&whenItem) },
		Assign:    func() { *l = List{whenItem} },
	}

	return marshal.UnmarshalOneOf(sliceCandidate, itemCandidate)
}

// Validate returns an error if any when clauses fail.
func (l *List) Validate(vars map[string]string) error {
	for _, w := range *l {
		if err := w.Validate(vars); err != nil {
			return err
		}
	}

	return nil
}

// Dependencies returns a list of options that are required explicitly.
// This does not include interpolations.
func (l *List) Dependencies() []string {
	if l == nil {
		return nil
	}

	// Use a map to prevent duplicates
	references := make(map[string]struct{})

	for _, w := range *l {
		for _, opt := range w.Dependencies() {
			references[opt] = struct{}{}
		}
	}

	options := make([]string, 0, len(references))
	for opt := range references {
		options = append(options, opt)
	}

	return options
}
