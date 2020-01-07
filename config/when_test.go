package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	yaml "gopkg.in/yaml.v2"
)

var unmarshalTests = []struct {
	desc     string
	input    string
	expected When
}{
	{
		"short notation",
		`foo`,
		createWhen(withWhenEqual("foo", "true")),
	},
	{
		"list short notation",
		`[foo, bar, baz]`,
		createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "true"), withWhenEqual("baz", "true")),
	},
	{
		"not-equal",
		`not-equal: {foo: bar}`,
		createWhen(withWhenNotEqual("foo", "bar")),
	},
	{
		"not-exists",
		`not-exists: file.txt`,
		createWhen(withWhenNotExists("file.txt")),
	},
	{
		"null environment",
		`environment: {foo: null}`,
		createWhen(withoutWhenEnv("foo")),
	},
	{
		"environment list with null",
		`environment: {foo: ["a", null, "b"]}`,
		createWhen(
			withWhenEnv("foo", "a"),
			withoutWhenEnv("foo"),
			withWhenEnv("foo", "b"),
		),
	},
}

func TestWhen_UnmarshalYAML(t *testing.T) {
	for _, tt := range unmarshalTests {
		w := When{}
		if err := yaml.UnmarshalStrict([]byte(tt.input), &w); err != nil {
			t.Errorf(
				`Unmarshaling %s: unexpected error: %s`,
				tt.desc, err,
			)
			continue
		}

		if !cmp.Equal(w, tt.expected) {
			t.Errorf("mismatch:\n%s", cmp.Diff(tt.expected, w))
		}
	}
}

var whenDepTests = []struct {
	when     When
	expected []string
}{
	{When{}, []string{}},

	// Equal
	{
		createWhen(withWhenEqual("foo", "true")),
		[]string{"foo"},
	},
	{
		createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "true")),
		[]string{"foo", "bar"},
	},

	// NotEqual
	{
		createWhen(withWhenNotEqual("foo", "true")),
		[]string{"foo"},
	},
	{
		createWhen(withWhenNotEqual("foo", "true"), withWhenNotEqual("bar", "true")),
		[]string{"foo", "bar"},
	},

	// Both
	{
		createWhen(withWhenEqual("foo", "true"), withWhenNotEqual("foo", "true")),
		[]string{"foo"},
	},
	{
		createWhen(withWhenEqual("foo", "true"), withWhenNotEqual("bar", "true")),
		[]string{"foo", "bar"},
	},
}

func TestWhen_Dependencies(t *testing.T) {
	for _, tt := range whenDepTests {
		actual := tt.when.Dependencies()
		if !equalUnordered(tt.expected, actual) {
			t.Errorf(
				"%+v.Dependencies(): expected %s, actual %s",
				tt.when, tt.expected, actual,
			)
		}
	}
}

var whenValidateTests = []struct {
	when      When
	options   map[string]string
	shouldErr bool
}{
	// Empty
	{When{}, nil, false},
	{When{}, map[string]string{"foo": "bar"}, false},

	// Command Clauses
	{createWhen(withWhenCommandSuccess), nil, false},
	{createWhen(withWhenCommandFailure), nil, true},
	{createWhen(withWhenCommandSuccess, withWhenCommandSuccess), nil, false},
	{createWhen(withWhenCommandSuccess, withWhenCommandFailure), nil, false},
	{createWhen(withWhenCommandFailure, withWhenCommandFailure), nil, true},

	// Exist Clauses
	{createWhen(withWhenExists("when_test.go")), nil, false},
	{createWhen(withWhenExists("fakefile")), nil, true},
	{createWhen(withWhenExists("fakefile"), withWhenExists("when_test.go")), nil, false},
	{createWhen(withWhenExists("when_test.go"), withWhenExists("fakefile")), nil, false},
	{createWhen(withWhenExists("fakefile"), withWhenExists("fakefile2")), nil, true},

	// Not Exist Clauses
	{createWhen(withWhenNotExists("when_test.go")), nil, true},
	{createWhen(withWhenNotExists("fakefile")), nil, false},
	{createWhen(withWhenNotExists("fakefile"), withWhenNotExists("when_test.go")), nil, false},
	{createWhen(withWhenNotExists("when_test.go"), withWhenNotExists("fakefile")), nil, false},
	{createWhen(withWhenNotExists("fakefile"), withWhenNotExists("fakefile2")), nil, false},
	{createWhen(withWhenNotExists("when.go"), withWhenNotExists("when_test.go")), nil, true},

	// OS Clauses
	{createWhen(withWhenOSSuccess), nil, false},
	{createWhen(withWhenOSFailure), nil, true},
	{createWhen(withWhenOSSuccess, withWhenOSFailure), nil, false},
	{createWhen(withWhenOSFailure, withWhenOSSuccess), nil, false},
	{createWhen(withWhenOSFailure, withWhenOSFailure), nil, true},

	// Environment Clauses
	{createWhen(withWhenEnvSuccess), nil, false},
	{createWhen(withoutWhenEnvSuccess), nil, false},
	{createWhen(withWhenEnvFailure), nil, true},
	{createWhen(withoutWhenEnvFailure), nil, true},
	{createWhen(withWhenEnvSuccess, withoutWhenEnvFailure), nil, false},
	{createWhen(withWhenEnvFailure, withoutWhenEnvSuccess), nil, false},
	{createWhen(withWhenEnvFailure, withoutWhenEnvFailure), nil, true},

	// Equal Clauses
	{
		createWhen(withWhenEqual("foo", "true")),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "false")),
		map[string]string{"foo": "true", "bar": "false"},
		false,
	},
	{
		createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "false")),
		map[string]string{"foo": "true", "bar": "true"},
		false,
	},
	{
		createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "false")),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(withWhenEqual("foo", "true")),
		map[string]string{"foo": "false"},
		true,
	},
	{
		createWhen(withWhenEqual("foo", "true")),
		map[string]string{},
		true,
	},
	{
		createWhen(withWhenEqual("foo", "true")),
		map[string]string{"bar": "true"},
		true,
	},
	{
		createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "false")),
		map[string]string{"bar": "true"},
		true,
	},

	// NotEqual Clauses
	{
		createWhen(withWhenNotEqual("foo", "true")),
		map[string]string{"foo": "true"},
		true,
	},
	{
		createWhen(withWhenNotEqual("foo", "true")),
		map[string]string{"foo": "false"},
		false,
	},
	{
		createWhen(withWhenNotEqual("foo", "true"), withWhenNotEqual("bar", "false")),
		map[string]string{"foo": "true", "bar": "true"},
		false,
	},
	{
		createWhen(withWhenNotEqual("foo", "true"), withWhenNotEqual("bar", "false")),
		map[string]string{"foo": "false"},
		false,
	},
	{
		createWhen(withWhenNotEqual("foo", "true")),
		map[string]string{},
		true,
	},
	{
		createWhen(withWhenNotEqual("foo", "true")),
		map[string]string{"bar": "true"},
		true,
	},
	{
		createWhen(withWhenNotEqual("foo", "true"), withWhenNotEqual("bar", "true")),
		map[string]string{"bar": "true"},
		true,
	},
	{
		createWhen(withWhenNotEqual("foo", "true"), withWhenNotEqual("bar", "true")),
		map[string]string{"foo": "false", "bar": "false"},
		false,
	},

	// Combinations
	{
		createWhen(
			withWhenCommandFailure,
			withWhenExists("fakefile"),
			withWhenNotExists("when_test.go"),
			withWhenOSFailure,
			withWhenEnvFailure,
			withWhenEqual("foo", "wrong"),
			withWhenNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		true,
	},
	{
		createWhen(
			withWhenCommandSuccess,
			withWhenExists("fakefile"),
			withWhenNotExists("when_test.go"),
			withWhenOSFailure,
			withWhenEnvFailure,
			withWhenEqual("foo", "wrong"),
			withWhenNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(
			withWhenCommandFailure,
			withWhenExists("when_test.go"),
			withWhenNotExists("when_test.go"),
			withWhenOSFailure,
			withWhenEnvFailure,
			withWhenEqual("foo", "wrong"),
			withWhenNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(
			withWhenCommandFailure,
			withWhenExists("fakefile"),
			withWhenNotExists("fakefile"),
			withWhenOSFailure,
			withWhenEnvFailure,
			withWhenEqual("foo", "wrong"),
			withWhenNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(
			withWhenCommandFailure,
			withWhenExists("fakefile"),
			withWhenNotExists("when_test.go"),
			withWhenOSSuccess,
			withWhenEnvFailure,
			withWhenEqual("foo", "wrong"),
			withWhenNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(
			withWhenCommandFailure,
			withWhenExists("fakefile"),
			withWhenNotExists("when_test.go"),
			withWhenOSFailure,
			withWhenEnvSuccess,
			withWhenEqual("foo", "wrong"),
			withWhenNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(
			withWhenCommandFailure,
			withWhenExists("fakefile"),
			withWhenNotExists("when_test.go"),
			withWhenOSFailure,
			withWhenEnvFailure,
			withWhenEqual("foo", "true"),
			withWhenNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(
			withWhenCommandFailure,
			withWhenExists("fakefile"),
			withWhenNotExists("when_test.go"),
			withWhenOSFailure,
			withWhenEnvFailure,
			withWhenEqual("foo", "wrong"),
			withWhenNotEqual("foo", "fake"),
		),
		map[string]string{"foo": "true"},
		false,
	},
}

func TestWhen_Validate(t *testing.T) {
	for _, tt := range whenValidateTests {
		err := tt.when.Validate(tt.options)
		didErr := err != nil
		if tt.shouldErr != didErr {
			t.Errorf(
				"%+v.Validate():\nexpected error: %t, got error: '%s'",
				tt.when, tt.shouldErr, err,
			)
		}
	}
}

var normalizetests = []struct {
	input    string
	expected string
}{
	{"nonsense", "nonsense"},
	{"darwin", "darwin"},
	{"Darwin", "darwin"},
	{"OSX", "darwin"},
	{"macOS", "darwin"},
	{"win", "windows"},
}

func TestNormalizeOS(t *testing.T) {
	for _, tt := range normalizetests {
		actual := normalizeOS(tt.input)
		if tt.expected != actual {
			t.Errorf(
				"normalizeOS(%s): expected: %s, actual: %s",
				tt.input, tt.expected, actual,
			)
		}
	}
}
