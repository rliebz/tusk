package when

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/rliebz/tusk/config/marshal"
	yaml "gopkg.in/yaml.v2"
)

// When defines the conditions for running a task.
type When struct {
	Command   marshal.StringList `yaml:",omitempty"`
	Exists    marshal.StringList `yaml:",omitempty"`
	NotExists marshal.StringList `yaml:"not-exists,omitempty"`
	OS        marshal.StringList `yaml:",omitempty"`

	Environment map[string]marshal.NullableStringList `yaml:",omitempty"`
	Equal       map[string]marshal.StringList         `yaml:",omitempty"`
	NotEqual    map[string]marshal.StringList         `yaml:"not-equal,omitempty"`
}

// UnmarshalYAML warns about deprecated features.
func (w *When) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var equal marshal.StringList
	slCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&equal) },
		Assign: func() {
			equalityMap := make(map[string]marshal.StringList, len(equal))
			for _, key := range equal {
				equalityMap[key] = marshal.StringList{"true"}
			}
			*w = When{Equal: equalityMap}
		},
	}

	type whenType When // Use new type to avoid recursion
	var whenItem whenType
	var ms yaml.MapSlice
	whenCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error {
			if err := unmarshal(&whenItem); err != nil {
				return err
			}

			if err := unmarshal(&ms); err != nil {
				return err
			}

			return nil
		},
		Assign: func() {
			*w = When(whenItem)

			// nil values on nullable string lists should be [nil], not []
			for _, clauseMS := range ms {
				if name, ok := clauseMS.Key.(string); !ok || name != "environment" {
					continue
				}

				for _, envMS := range clauseMS.Value.(yaml.MapSlice) {
					envVar := envMS.Key.(string)

					if envMS.Value == nil {
						w.Environment[envVar] = marshal.NullableStringList{nil}
					}
				}
			}
		},
	}

	return marshal.UnmarshalOneOf(slCandidate, whenCandidate)
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
		w.validateNotExists(),
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

func (w *When) validateNotExists() error {
	if len(w.NotExists) == 0 {
		return newUnspecifiedError("not-exists")
	}

	for _, f := range w.NotExists {
		if _, err := os.Stat(f); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		return newCondFailErrorf("file exists, but shouldn't: %s", w.NotExists)
	}

	return nil
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

func normalizeOS(name string) string {
	lower := strings.ToLower(name)

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
	_, err := exec.Command("sh", "-c", command).Output() // nolint: gosec
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
