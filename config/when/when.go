package when

import (
	"errors"
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

	Environment map[string]marshal.NullableStringList `yaml:",omitempty"`
	Equal       map[string]marshal.StringList         `yaml:",omitempty"`
	NotEqual    map[string]marshal.StringList         `yaml:"not-equal,omitempty"`
	// Deprecated: Use `not-equal` instead of `not_equal`
	NotEqualDeprecated map[string]marshal.StringList `yaml:"not_equal,omitempty"`
}

func (w When) String() string {
	output := make([]string, 0, 6)
	if len(w.Command) > 0 {
		output = append(output, fmt.Sprintf("command:%s", w.Command))
	}
	if len(w.Exists) > 0 {
		output = append(output, fmt.Sprintf("exists:%s", w.Exists))
	}
	if len(w.OS) > 0 {
		output = append(output, fmt.Sprintf("os:%s", w.OS))
	}
	if len(w.Environment) > 0 {
		output = append(output, "environment:"+sprintNullableMap(w.Environment))
	}
	if len(w.Equal) > 0 {
		output = append(output, "equal:"+sprintMap(w.Equal))
	}
	if len(w.NotEqual) > 0 {
		output = append(output, "not-equal:"+sprintMap(w.NotEqual))
	}

	return "When{" + strings.Join(output, ",") + "}"
}

func sprintNullableMap(m map[string]marshal.NullableStringList) string {
	output := make([]string, 0, len(m))
	for k, v := range m {
		list := make([]string, 0, len(v))
		for _, item := range v {
			if item == nil {
				list = append(list, "nil")
			} else {
				list = append(list, *item)
			}
		}
		listString := "[" + strings.Join(list, ",") + "]"
		output = append(output, fmt.Sprintf("%s:%s", k, listString))
	}

	return "{" + strings.Join(output, ",") + "}"

}

func sprintMap(m map[string]marshal.StringList) string {
	output := make([]string, 0, len(m))
	for k, v := range m {
		listString := "[" + strings.Join(v, ",") + "]"
		output = append(output, fmt.Sprintf("%s:%s", k, listString))
	}

	return "{" + strings.Join(output, ",") + "}"
}

// UnmarshalYAML warns about deprecated features.
func (w *When) UnmarshalYAML(unmarshal func(interface{}) error) error {

	type whenType When // Use new type to avoid recursion
	if err := unmarshal((*whenType)(w)); err != nil {
		return err
	}

	if len(w.NotEqualDeprecated) > 0 {
		ui.Deprecate("The `not_equal` clause has been renamed to `not-equal`")
		if len(w.NotEqual) > 0 {
			return errors.New(
				"both `not_equal` and `not-equal` are defined in a single `when` clause",
			)
		}

		w.NotEqual = w.NotEqualDeprecated
		w.NotEqualDeprecated = nil
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

	return validateAny(
		w.validateOS(),
		w.validateEqual(vars),
		w.validateNotEqual(vars),
		w.validateEnv(),
		w.validateExists(),
		w.validateCommand(),
	)
}

// TODO: Should this be done in parallel?
func validateAny(errs ...error) error {
	var errOutput error
	for _, err := range errs {
		if err == nil {
			return nil
		}

		if errOutput == nil && !IsUnspecifiedClause(err) {
			errOutput = err
		}
	}

	return errOutput
}

func (w *When) validateCommand() error {
	if len(w.Command) == 0 {
		return newUnspecifiedError("command")
	}

	for _, command := range w.Command {
		if err := testCommand(command); err == nil {
			return nil
		}
	}

	return newCondFailErrorf("no commands exited successfully")
}

func (w *When) validateExists() error {
	if len(w.Exists) == 0 {
		return newUnspecifiedError("exists")
	}

	for _, f := range w.Exists {
		if _, err := os.Stat(f); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			continue
		}

		return nil
	}

	return newCondFailErrorf("no required file existed: %s", w.Exists)
}

func (w *When) validateOS() error {
	if len(w.OS) == 0 {
		return newUnspecifiedError("os")
	}

	return validateOneOf(
		"current OS", runtime.GOOS, w.OS,
		func(expected, actual string) bool {
			return normalizeOS(expected) == actual
		},
	)
}

func (w *When) validateEnv() error {
	if len(w.Environment) == 0 {
		return newUnspecifiedError("env")
	}

	for varName, values := range w.Environment {
		stringValues := make([]string, 0, len(values))
		for _, value := range values {
			if value != nil {
				stringValues = append(stringValues, *value)
			}
		}

		isNullAllowed := len(values) != len(stringValues)

		actual, ok := os.LookupEnv(varName)
		if !ok {
			if isNullAllowed {
				return nil
			}

			continue
		}

		if err := validateOneOf(
			fmt.Sprintf("environment variable %s", varName),
			actual,
			stringValues,
			func(a, b string) bool { return a == b },
		); err == nil {
			return nil
		}
	}

	return newCondFailError("no environment variables matched")
}

func (w *When) validateEqual(vars map[string]string) error {
	if len(w.Equal) == 0 {
		return newUnspecifiedError("equal")
	}

	return validateEquality(vars, w.Equal, func(a, b string) bool {
		return a == b
	})
}

func (w *When) validateNotEqual(vars map[string]string) error {
	if len(w.NotEqual) == 0 {
		return newUnspecifiedError("not-equal")
	}

	return validateEquality(vars, w.NotEqual, func(a, b string) bool {
		return a != b
	})
}

func validateOneOf(
	desc, value string, required []string, compare func(string, string) bool,
) error {
	for _, expected := range required {
		if compare(expected, value) {
			return nil
		}
	}

	return newCondFailErrorf("%s (%s) not listed in %v", desc, value, required)
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
			continue
		}

		if err := validateOneOf(
			fmt.Sprintf(`option "%s"`, optionName),
			actual,
			values,
			compare,
		); err == nil {
			return nil
		}
	}

	return newCondFailError("no options matched")
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

// TODO: Remove deprecations
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
          ...`,
			example1, example2),
	)
}
