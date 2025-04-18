package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/rliebz/tusk/marshal"
)

// When defines the conditions for running a task.
type When struct {
	Command   marshal.Slice[string] `yaml:",omitempty"`
	Exists    marshal.Slice[string] `yaml:",omitempty"`
	NotExists marshal.Slice[string] `yaml:"not-exists,omitempty"`
	OS        marshal.Slice[string] `yaml:",omitempty"`

	Environment map[string]marshal.Slice[*string] `yaml:",omitempty"`
	Equal       map[string]marshal.Slice[string]  `yaml:",omitempty"`
	NotEqual    map[string]marshal.Slice[string]  `yaml:"not-equal,omitempty"`
}

// UnmarshalYAML warns about deprecated features.
func (w *When) UnmarshalYAML(unmarshal func(any) error) error {
	var equal marshal.Slice[string]
	slCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&equal) },
		Assign: func() {
			equalityMap := make(map[string]marshal.Slice[string], len(equal))
			for _, key := range equal {
				equalityMap[key] = marshal.Slice[string]{"true"}
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
			fixNilEnvironment(w, ms)
		},
	}

	return marshal.UnmarshalOneOf(slCandidate, whenCandidate)
}

// fixNilEnvironment replaces a single nil specified in a yaml configuration as
// a list of nil, which is the more logical interpretation of the value in this
// situation.
func fixNilEnvironment(w *When, ms yaml.MapSlice) {
	for _, clauseMS := range ms {
		if name, ok := clauseMS.Key.(string); !ok || name != "environment" {
			continue
		}

		for _, envMS := range clauseMS.Value.(yaml.MapSlice) {
			envVar := envMS.Key.(string)

			if envMS.Value == nil {
				w.Environment[envVar] = marshal.Slice[*string]{nil}
			}
		}
	}
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
func (w *When) Validate(ctx Context, vars map[string]string) error {
	if w == nil {
		return nil
	}

	return validateAny(
		w.validateOS(),
		w.validateEqual(vars),
		w.validateNotEqual(vars),
		w.validateEnv(),
		w.validateExists(ctx),
		w.validateNotExists(ctx),
		w.validateCommand(ctx),
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

func (w *When) validateCommand(ctx Context) error {
	if len(w.Command) == 0 {
		return newUnspecifiedError("command")
	}

	for _, command := range w.Command {
		if err := testCommand(ctx, command); err == nil {
			return nil
		}
	}

	return newCondFailErrorf("no commands exited successfully")
}

func (w *When) validateExists(ctx Context) error {
	if len(w.Exists) == 0 {
		return newUnspecifiedError("exists")
	}

	for _, f := range w.Exists {
		if _, err := os.Stat(filepath.Join(ctx.Dir(), f)); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			continue
		}

		return nil
	}

	return newCondFailErrorf("no required file exists: %s", w.Exists)
}

func (w *When) validateNotExists(ctx Context) error {
	if len(w.NotExists) == 0 {
		return newUnspecifiedError("not-exists")
	}

	for _, f := range w.NotExists {
		if _, err := os.Stat(filepath.Join(ctx.Dir(), f)); err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
	}

	return newCondFailErrorf("all files exist: %s", w.NotExists)
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
		if w.isEnvVarValid(varName, values) {
			return nil
		}
	}

	return newCondFailError("no environment variables matched")
}

func (w *When) isEnvVarValid(varName string, values marshal.Slice[*string]) bool {
	stringValues := make([]string, 0, len(values))
	for _, value := range values {
		if value != nil {
			stringValues = append(stringValues, *value)
		}
	}

	isNullAllowed := len(values) != len(stringValues)

	actual, ok := os.LookupEnv(varName)
	if !ok {
		return isNullAllowed
	}

	err := validateOneOf(
		"environment variable "+varName,
		actual,
		stringValues,
		func(a, b string) bool { return a == b },
	)
	return err == nil
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

func testCommand(ctx Context, command string) error {
	cmd := newCmd(ctx, command)
	_, err := cmd.Output()
	return err
}

func validateEquality(
	options map[string]string,
	cases map[string]marshal.Slice[string],
	compare func(string, string) bool,
) error {
	for optionName, values := range cases {
		actual, ok := options[optionName]
		if !ok {
			continue
		}

		if err := validateOneOf(
			fmt.Sprintf("option %q", optionName),
			actual,
			values,
			compare,
		); err == nil {
			return nil
		}
	}

	return newCondFailError("no options matched")
}

// WhenList is a list of when items with custom yaml unmarshaling.
type WhenList marshal.Slice[When]

// UnmarshalYAML allows single items to be used as lists.
func (l *WhenList) UnmarshalYAML(unmarshal func(any) error) error {
	return (*marshal.Slice[When])(l).UnmarshalYAML(unmarshal)
}

// Validate returns an error if any when clauses fail.
func (l *WhenList) Validate(ctx Context, vars map[string]string) error {
	if l == nil {
		return nil
	}

	for _, w := range *l {
		if err := w.Validate(ctx, vars); err != nil {
			return err
		}
	}

	return nil
}

// Dependencies returns a list of options that are required explicitly.
// This does not include interpolations.
func (l *WhenList) Dependencies() []string {
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
