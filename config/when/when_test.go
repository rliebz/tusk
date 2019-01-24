package when

import (
	"reflect"
	"testing"

	"github.com/rliebz/tusk/config/marshal"
	yaml "gopkg.in/yaml.v2"
)

var s = "string"
var stringTests = []struct {
	desc     string
	input    When
	expected string
}{
	{
		"no fields",
		When{},
		"When{}",
	},
	{
		"one list",
		When{Command: marshal.StringList{"foo"}},
		"When{command:[foo]}",
	},
	{
		"one nullable map",
		When{
			Environment: map[string]marshal.NullableStringList{"foo": {nil, &s}},
		},
		"When{environment:{foo:[nil,string]}}",
	},
	{
		"one map",
		When{
			Equal: map[string]marshal.StringList{"foo": {"bar", "baz"}},
		},
		"When{equal:{foo:[bar,baz]}}",
	},
	{
		"all lists",
		When{
			Command: marshal.StringList{"foo"},
			Exists:  marshal.StringList{"bar"},
			OS:      marshal.StringList{"baz"},
		},
		"When{command:[foo],exists:[bar],os:[baz]}",
	},
	{
		"all maps",
		When{
			Environment: map[string]marshal.NullableStringList{"env": {nil, &s}},
			Equal:       map[string]marshal.StringList{"foo": {"bar", "baz"}},
			NotEqual:    map[string]marshal.StringList{"a": {"b", "c"}},
		},
		"When{environment:{env:[nil,string]},equal:{foo:[bar,baz]},not-equal:{a:[b,c]}}",
	},
}

func TestWhen_String(t *testing.T) {
	for _, tt := range stringTests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := tt.input.String()
			if tt.expected != actual {
				t.Errorf("want %q; got %q", tt.expected, actual)
			}
		})
	}
}

var unmarshalTests = []struct {
	desc     string
	input    string
	expected When
}{
	{
		"short notation",
		`foo`,
		Create(WithEqual("foo", "true")),
	},
	{
		"list short notation",
		`[foo, bar]`,
		Create(WithEqual("foo", "true"), WithEqual("bar", "true")),
	},
	{
		"not-equal",
		`not-equal: {foo: bar}`,
		Create(WithNotEqual("foo", "bar")),
	},
	{
		"null environment",
		`environment: {foo: null}`,
		Create(WithoutEnv("foo")),
	},
	{
		"environment list with null",
		`environment: {foo: ["a", null, "b"]}`,
		Create(
			WithEnv("foo", "a"),
			WithoutEnv("foo"),
			WithEnv("foo", "b"),
		),
	},
}

func TestWhen_UnmarshalYAML(t *testing.T) {
	for _, tt := range unmarshalTests {
		w := When{}
		if err := yaml.Unmarshal([]byte(tt.input), &w); err != nil {
			t.Errorf(
				`Unmarshalling %s: unexpected error: %s`,
				tt.desc, err,
			)
			continue
		}

		expected := tt.expected.String()
		actual := w.String()

		if expected != actual {
			t.Errorf("want %q, got %q", expected, actual)
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
		Create(WithEqual("foo", "true")),
		[]string{"foo"},
	},
	{
		Create(WithEqual("foo", "true"), WithEqual("bar", "true")),
		[]string{"foo", "bar"},
	},

	// NotEqual
	{
		Create(WithNotEqual("foo", "true")),
		[]string{"foo"},
	},
	{
		Create(WithNotEqual("foo", "true"), WithNotEqual("bar", "true")),
		[]string{"foo", "bar"},
	},

	// Both
	{
		Create(WithEqual("foo", "true"), WithNotEqual("foo", "true")),
		[]string{"foo"},
	},
	{
		Create(WithEqual("foo", "true"), WithNotEqual("bar", "true")),
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

func equalUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Since this list is unordered, convert to maps
	aMap := make(map[string]interface{})
	for _, val := range a {
		aMap[val] = struct{}{}
	}

	bMap := make(map[string]interface{})
	for _, val := range b {
		bMap[val] = struct{}{}
	}

	return reflect.DeepEqual(aMap, bMap)
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
	{Create(WithCommandSuccess), nil, false},
	{Create(WithCommandFailure), nil, true},
	{Create(WithCommandSuccess, WithCommandSuccess), nil, false},
	{Create(WithCommandSuccess, WithCommandFailure), nil, false},
	{Create(WithCommandFailure, WithCommandFailure), nil, true},

	// Exist Clauses
	{Create(WithExists("when_test.go")), nil, false},
	{Create(WithExists("fakefile")), nil, true},
	{Create(WithExists("fakefile"), WithExists("when_test.go")), nil, false},
	{Create(WithExists("when_test.go"), WithExists("fakefile")), nil, false},
	{Create(WithExists("fakefile"), WithExists("fakefile2")), nil, true},

	// OS Clauses
	{Create(WithOSSuccess), nil, false},
	{Create(WithOSFailure), nil, true},
	{Create(WithOSSuccess, WithOSFailure), nil, false},
	{Create(WithOSFailure, WithOSSuccess), nil, false},
	{Create(WithOSFailure, WithOSFailure), nil, true},

	// Environment Clauses
	{Create(WithEnvSuccess), nil, false},
	{Create(WithoutEnvSuccess), nil, false},
	{Create(WithEnvFailure), nil, true},
	{Create(WithoutEnvFailure), nil, true},
	{Create(WithEnvSuccess, WithoutEnvFailure), nil, false},
	{Create(WithEnvFailure, WithoutEnvSuccess), nil, false},
	{Create(WithEnvFailure, WithoutEnvFailure), nil, true},

	// Equal Clauses
	{
		Create(WithEqual("foo", "true")),
		map[string]string{"foo": "true"},
		false,
	},
	{
		Create(WithEqual("foo", "true"), WithEqual("bar", "false")),
		map[string]string{"foo": "true", "bar": "false"},
		false,
	},
	{
		Create(WithEqual("foo", "true"), WithEqual("bar", "false")),
		map[string]string{"foo": "true", "bar": "true"},
		false,
	},
	{
		Create(WithEqual("foo", "true"), WithEqual("bar", "false")),
		map[string]string{"foo": "true"},
		false,
	},
	{
		Create(WithEqual("foo", "true")),
		map[string]string{"foo": "false"},
		true,
	},
	{
		Create(WithEqual("foo", "true")),
		map[string]string{},
		true,
	},
	{
		Create(WithEqual("foo", "true")),
		map[string]string{"bar": "true"},
		true,
	},
	{
		Create(WithEqual("foo", "true"), WithEqual("bar", "false")),
		map[string]string{"bar": "true"},
		true,
	},

	// NotEqual Clauses
	{
		Create(WithNotEqual("foo", "true")),
		map[string]string{"foo": "true"},
		true,
	},
	{
		Create(WithNotEqual("foo", "true")),
		map[string]string{"foo": "false"},
		false,
	},
	{
		Create(WithNotEqual("foo", "true"), WithNotEqual("bar", "false")),
		map[string]string{"foo": "true", "bar": "true"},
		false,
	},
	{
		Create(WithNotEqual("foo", "true"), WithNotEqual("bar", "false")),
		map[string]string{"foo": "false"},
		false,
	},
	{
		Create(WithNotEqual("foo", "true")),
		map[string]string{},
		true,
	},
	{
		Create(WithNotEqual("foo", "true")),
		map[string]string{"bar": "true"},
		true,
	},
	{
		Create(WithNotEqual("foo", "true"), WithNotEqual("bar", "true")),
		map[string]string{"bar": "true"},
		true,
	},
	{
		Create(WithNotEqual("foo", "true"), WithNotEqual("bar", "true")),
		map[string]string{"foo": "false", "bar": "false"},
		false,
	},

	// Combinations
	{
		Create(
			WithCommandFailure,
			WithExists("fakefile"),
			WithOSFailure,
			WithEnvFailure,
			WithEqual("foo", "wrong"),
			WithNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		true,
	},
	{
		Create(
			WithCommandSuccess,
			WithExists("fakefile"),
			WithOSFailure,
			WithEnvFailure,
			WithEqual("foo", "wrong"),
			WithNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		Create(
			WithCommandFailure,
			WithExists("when_test.go"),
			WithOSFailure,
			WithEnvFailure,
			WithEqual("foo", "wrong"),
			WithNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		Create(
			WithCommandFailure,
			WithExists("fakefile"),
			WithOSSuccess,
			WithEnvFailure,
			WithEqual("foo", "wrong"),
			WithNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		Create(
			WithCommandFailure,
			WithExists("fakefile"),
			WithOSFailure,
			WithEnvSuccess,
			WithEqual("foo", "wrong"),
			WithNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		Create(
			WithCommandFailure,
			WithExists("fakefile"),
			WithOSFailure,
			WithEnvFailure,
			WithEqual("foo", "true"),
			WithNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		Create(
			WithCommandFailure,
			WithExists("fakefile"),
			WithOSFailure,
			WithEnvFailure,
			WithEqual("foo", "wrong"),
			WithNotEqual("foo", "fake"),
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
