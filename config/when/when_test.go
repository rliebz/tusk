package when

import (
	"reflect"
	"testing"
)

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
	{Create(WithCommandSuccess, WithCommandFailure), nil, true},

	// Exist Clauses
	{Create(WithExists("when_test.go")), nil, false},
	{Create(WithExists("fakefile")), nil, true},
	{Create(WithExists("fakefile"), WithExists("when_test.go")), nil, true},

	// OS Clauses
	{Create(WithOSSuccess), nil, false},
	{Create(WithOSFailure), nil, true},
	{Create(WithOSSuccess, WithOSFailure), nil, false},
	{Create(WithOSFailure, WithOSSuccess), nil, false},

	// Environment Clauses
	{Create(WithEnvSuccess), nil, false},
	{Create(WithoutEnvSuccess), nil, false},
	{Create(WithEnvFailure), nil, true},
	{Create(WithoutEnvFailure), nil, true},
	{Create(WithEnvSuccess, WithoutEnvFailure), nil, true},
	{Create(WithEnvFailure, WithoutEnvSuccess), nil, true},

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
