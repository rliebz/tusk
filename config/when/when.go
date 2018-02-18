package when

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/rliebz/tusk/config/marshal"
	"github.com/rliebz/tusk/ui"
)

// When defines the conditions for running a task.
type When struct {
	Command marshal.StringList `yaml:",omitempty"`
	Exists  marshal.StringList `yaml:",omitempty"`
	OS      marshal.StringList `yaml:",omitempty"`

	Environment map[string]*string            `yaml:",omitempty"`
	Equal       map[string]marshal.StringList `yaml:",omitempty"`
	NotEqual    map[string]marshal.StringList `yaml:"not_equal,omitempty"`
}

// UnmarshalYAML warns about deprecated features.
func (w *When) UnmarshalYAML(unmarshal func(interface{}) error) error {

	type whenType When // Use new type to avoid recursion
	if err := unmarshal((*whenType)(w)); err != nil {
		return err
	}

	warnDeprecations(w)

	return nil
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

	if err := w.validateExists(); err != nil {
		return err
	}

	if err := w.validateOS(); err != nil {
		return err
	}

	if err := w.validateEnv(); err != nil {
		return err
	}

	if err := w.validateEqual(vars); err != nil {
		return err
	}

	if err := w.validateNotEqual(vars); err != nil {
		return err
	}

	if err := w.validateCommand(); err != nil {
		return err
	}

	return nil
}

func (w *When) validateCommand() error {
	for _, command := range w.Command {
		if err := testCommand(command); err != nil {
			return newCondFailErrorf(`test failed: %s`, command)
		}
	}

	return nil
}

func (w *When) validateExists() error {
	for _, f := range w.Exists {
		if _, err := os.Stat(f); err != nil {
			if os.IsNotExist(err) {
				return newCondFailErrorf(`file "%s" does not exist`, f)
			}
			return err
		}
	}

	return nil
}

func (w *When) validateOS() error {
	return validateOneOf(
		"current OS", runtime.GOOS, w.OS,
		func(expected, actual string) bool {
			return normalizeOS(expected) == actual
		},
	)
}

func (w *When) validateEnv() error {
	for varName, expected := range w.Environment {
		actual, ok := os.LookupEnv(varName)
		if expected == nil {
			if ok {
				return newCondFailErrorf(
					`environment variable %s ("%s") must not be set`,
					varName, actual,
				)
			}

			continue
		}

		if *expected != actual {
			return newCondFailErrorf(
				`environment variable %s ("%s") does not match expected value (%s)`,
				varName, actual, expected,
			)
		}
	}

	return nil
}

func (w *When) validateEqual(vars map[string]string) error {
	return validateEquality(vars, w.Equal, func(a, b string) bool {
		return a == b
	})
}

func (w *When) validateNotEqual(vars map[string]string) error {
	return validateEquality(vars, w.NotEqual, func(a, b string) bool {
		return a != b
	})
}

// nolint: unparam
func validateOneOf(
	desc, value string, required []string, compare func(string, string) bool,
) error {
	if len(required) == 0 {
		return nil
	}

	for _, expected := range required {
		if compare(expected, value) {
			return nil
		}
	}

	return newCondFailErrorf(`%s (%s) not listed in %v`, desc, value, required)
}

// validateAllOf is a stand-in for validateOneOf to be backward compatible.
// Behavior will change in an upcoming major version, as this behavior has
// been deprecated.
func validateAllOf(
	desc, value string, required []string, compare func(string, string) bool,
) error {
	if len(required) == 0 {
		return nil
	}

	for _, expected := range required {
		if !compare(expected, value) {
			return newCondFailErrorf(
				`%s (%s) does not match expected value (%s)`,
				desc, value, expected,
			)
		}
	}

	return nil
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

	for optionName, values := range cases {
		actual, ok := options[optionName]
		if !ok {
			return newCondFailErrorf(`option "%s" not defined`, optionName)
		}

		if err := validateAllOf(
			fmt.Sprintf(`option "%s"`, optionName),
			actual,
			values,
			compare,
		); err != nil {
			return err
		}
	}

	return nil
}

func warnDeprecations(w *When) {
	warnMultiClauseDeprecation(w)

	warnListDeprecation(w.Command, "command")
	warnListDeprecation(w.Exists, "exists")

	for _, l := range w.Equal {
		warnListDeprecation(l, "equal")
	}

	for _, l := range w.NotEqual {
		warnListDeprecation(l, "equal")
	}
}

func warnMultiClauseDeprecation(w *When) {
	var clausesUsed []string

	if len(w.Command) > 0 {
		clausesUsed = append(clausesUsed, "command")
	}

	if len(w.Exists) > 0 {
		clausesUsed = append(clausesUsed, "exists")
	}

	if len(w.OS) > 0 {
		clausesUsed = append(clausesUsed, "os")
	}

	if len(w.Environment) > 0 {
		clausesUsed = append(clausesUsed, "environment")
	}

	if len(w.Equal) > 0 {
		clausesUsed = append(clausesUsed, "equal")
	}

	if len(w.NotEqual) > 0 {
		clausesUsed = append(clausesUsed, "not_equal")
	}

	if len(clausesUsed) > 1 {
		deprecateWhenBehavior(
			"Using multiple checks",
			clausesUsed[0], clausesUsed[1],
		)
	}
}

func warnListDeprecation(list marshal.StringList, field string) {
	if len(list) > 1 {
		deprecateWhenBehavior(
			fmt.Sprintf("Multiple values for `%s`", field),
			field, field,
		)
	}
}

func deprecateWhenBehavior(behavior, example1, example2 string) {
	ui.Deprecate(
		fmt.Sprintf("%s in `when` clauses has been deprecated", behavior),
		"The behavior will change in a future release",
		fmt.Sprintf(`Use multiple when clauses for multiple requirements instead

        when:
          - %s: ...
          - %s: ...
          ...
`,
			example1, example2),
	)
}
