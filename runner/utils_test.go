package runner

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"

	"github.com/rliebz/tusk/marshal"
)

// createOption creates a custom option for testing purposes.
func createOption(operators ...func(o *Option)) *Option {
	o := Option{}

	for _, f := range operators {
		f(&o)
	}

	return &o
}

// withOptionName returns an operator that adds a name to an option.
func withOptionName(name string) func(o *Option) {
	return func(o *Option) {
		o.Name = name
	}
}

// withOptionDependency returns an operator that adds a dependency to an option.
func withOptionDependency(name string) func(o *Option) {
	return func(o *Option) {
		o.DefaultValues = append(
			o.DefaultValues,
			Value{
				Value: fmt.Sprintf("${%s}", name),
			},
		)
	}
}

// withOptionWhenDependency returns an operator that adds a when dependency to an option.
func withOptionWhenDependency(name string) func(o *Option) {
	return func(o *Option) {
		o.DefaultValues = append(
			o.DefaultValues,
			Value{When: WhenList{createWhen(withWhenEqual(name, "true"))}},
		)
	}
}

// whenTrue is a When that always evaluates to true.
var whenTrue = When{}

// whenFalse is a When that always evaluates to false.
var whenFalse = When{OS: marshal.StringList{"fake"}}

// createWhen creates a custom when for testing purposes.
func createWhen(operators ...func(w *When)) When {
	w := When{}

	for _, f := range operators {
		f(&w)
	}

	return w
}

// withWhenCommand returns an operator that runs a given command
func withWhenCommand(command string) func(w *When) {
	return func(w *When) {
		w.Command = append(w.Command, command)
	}
}

// withWhenCommandSuccess is an operator that includes a successful command.
var withWhenCommandSuccess = func(w *When) {
	w.Command = append(w.Command, "test 1 = 1")
}

// withWhenCommandFailure is an operator that includes a failed command.
var withWhenCommandFailure = func(w *When) {
	w.Command = append(w.Command, "test 0 = 1")
}

// withWhenExists returns an operator that requires a file to exist.
func withWhenExists(filename string) func(w *When) {
	return func(w *When) {
		w.Exists = append(w.Exists, filename)
	}
}

// withWhenNotExists returns an operator that requires a file to not exist.
func withWhenNotExists(filename string) func(w *When) {
	return func(w *When) {
		w.NotExists = append(w.NotExists, filename)
	}
}

// withWhenOS returns an operator that requires an arbitrary OS.
func withWhenOS(name string) func(w *When) {
	return func(w *When) {
		w.OS = append(w.OS, name)
	}
}

// withWhenOSSuccess is an operator that requires the current OS.
var withWhenOSSuccess = func(w *When) {
	w.OS = append(w.OS, runtime.GOOS)
}

// withWhenOSFailure is an operator that requires the wrong OS.
var withWhenOSFailure = func(w *When) {
	w.OS = append(w.OS, "fake")
}

// withWhenEnv returns an operator that requires an env var to be set.
func withWhenEnv(key, value string) func(w *When) {
	return func(w *When) {
		ensureEnv(w)
		w.Environment[key] = append(w.Environment[key], &value)
	}
}

// withoutWhenEnv returns an operator that requires an env var to be unset.
func withoutWhenEnv(key string) func(w *When) {
	return func(w *When) {
		ensureEnv(w)
		w.Environment[key] = append(w.Environment[key], nil)
	}
}

// withWhenEnvSuccess is an operator that requires a set environment variable.
var withWhenEnvSuccess = func(w *When) {
	ensureEnv(w)
	key := randomString()
	value := randomString()
	os.Setenv(key, value) //nolint: errcheck
	w.Environment[key] = append(w.Environment[key], &value)
}

// withWhenEnvFailure is an operator that requires a set environment variable.
var withWhenEnvFailure = func(w *When) {
	ensureEnv(w)
	key := randomString()
	value := randomString()
	w.Environment[key] = append(w.Environment[key], &value)
}

// withoutWhenEnvSuccess is an operator that requires an unset environment variable.
var withoutWhenEnvSuccess = func(w *When) {
	ensureEnv(w)
	key := randomString()
	w.Environment[key] = append(w.Environment[key], nil)
}

// withoutWhenEnvFailure is an operator that requires an unset environment variable.
var withoutWhenEnvFailure = func(w *When) {
	ensureEnv(w)
	key := randomString()
	value := randomString()
	os.Setenv(key, value) //nolint: errcheck
	w.Environment[key] = append(w.Environment[key], nil)
}

func randomString() string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ_")
	a := make([]rune, 20)
	for i := range a {
		a[i] = letters[rand.Intn(len(letters))]
	}
	return string(a)
}

func ensureEnv(w *When) {
	if w.Environment == nil {
		w.Environment = make(map[string]marshal.NullableStringList)
	}
}

// withWhenEqual returns an operator that requires the key to equal the value.
func withWhenEqual(key, value string) func(w *When) {
	return func(w *When) {
		if w.Equal == nil {
			w.Equal = make(map[string]marshal.StringList)
		}

		w.Equal[key] = append(w.Equal[key], value)
	}
}

// withWhenNotEqual returns an operator that requires the key to not equal the value.
func withWhenNotEqual(key, value string) func(w *When) {
	return func(w *When) {
		if w.NotEqual == nil {
			w.NotEqual = make(map[string]marshal.StringList)
		}
		w.NotEqual[key] = append(w.NotEqual[key], value)
	}
}
