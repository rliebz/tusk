package when

import (
	"math/rand"
	"os"
	"runtime"

	"github.com/rliebz/tusk/config/marshal"
)

// True is a when.When that always evaluates to true.
var True = When{}

// False is a when.When that always evaluates to false.
var False = When{OS: marshal.StringList{"fake"}}

// Create creates a custom when for testing purposes.
func Create(operators ...func(w *When)) When {
	w := When{}

	for _, f := range operators {
		f(&w)
	}

	return w
}

// WithCommand returns an operator that runs a given command
func WithCommand(command string) func(w *When) {
	return func(w *When) {
		w.Command = append(w.Command, command)
	}
}

// WithCommandSuccess is an operator that includes a successful command.
var WithCommandSuccess = func(w *When) {
	w.Command = append(w.Command, "test 1 = 1")
}

// WithCommandFailure is an operator that includes a failed command.
var WithCommandFailure = func(w *When) {
	w.Command = append(w.Command, "test 0 = 1")
}

// WithExists returns an operator that requires a file to exist.
func WithExists(filename string) func(w *When) {
	return func(w *When) {
		w.Exists = append(w.Exists, filename)
	}
}

// WithOS returns an operator that requires an arbitrary OS.
func WithOS(name string) func(w *When) {
	return func(w *When) {
		w.OS = append(w.OS, name)
	}
}

// WithOSSuccess is an operator that requires the current OS.
var WithOSSuccess = func(w *When) {
	w.OS = append(w.OS, runtime.GOOS)
}

// WithOSFailure is an operator that requires the wrong OS.
var WithOSFailure = func(w *When) {
	w.OS = append(w.OS, "fake")
}

// WithEnv returns an operator that requires an env var to be set.
func WithEnv(key, value string) func(w *When) {
	return func(w *When) {
		ensureEnv(w)
		w.Environment[key] = append(w.Environment[key], &value)
	}
}

// WithoutEnv returns an operator that requires an env var to be unset.
func WithoutEnv(key string) func(w *When) {
	return func(w *When) {
		ensureEnv(w)
		w.Environment[key] = append(w.Environment[key], nil)
	}
}

// WithEnvSuccess is an operator that requires a set environment variable.
var WithEnvSuccess = func(w *When) {
	ensureEnv(w)
	key := randomString()
	value := randomString()
	os.Setenv(key, value) // nolint: errcheck
	w.Environment[key] = append(w.Environment[key], &value)
}

// WithEnvFailure is an operator that requires a set environment variable.
var WithEnvFailure = func(w *When) {
	ensureEnv(w)
	key := randomString()
	value := randomString()
	w.Environment[key] = append(w.Environment[key], &value)
}

// WithoutEnvSuccess is an operator that requires an unset environment variable.
var WithoutEnvSuccess = func(w *When) {
	ensureEnv(w)
	key := randomString()
	w.Environment[key] = append(w.Environment[key], nil)
}

// WithoutEnvFailure is an operator that requires an unset environment variable.
var WithoutEnvFailure = func(w *When) {
	ensureEnv(w)
	key := randomString()
	value := randomString()
	os.Setenv(key, value) // nolint: errcheck
	w.Environment[key] = append(w.Environment[key], nil)
}

func randomString() string {
	var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ_")
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

// WithEqual returns an operator that requires the key to equal the value.
func WithEqual(key, value string) func(w *When) {
	return func(w *When) {
		if w.Equal == nil {
			w.Equal = make(map[string]marshal.StringList)
		}

		w.Equal[key] = append(w.Equal[key], value)
	}
}

// WithNotEqual returns an operator that requires the key to not equal the value.
func WithNotEqual(key, value string) func(w *When) {
	return func(w *When) {
		if w.NotEqual == nil {
			w.NotEqual = make(map[string]marshal.StringList)
		}
		w.NotEqual[key] = append(w.NotEqual[key], value)
	}
}
