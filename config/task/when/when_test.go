package when

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/rliebz/tusk/config/task/marshal"
)

// Define convenience aliases.
type eqMap = map[string]marshal.StringList
type sl = marshal.StringList

var dependenciestests = []struct {
	when     *When
	expected []string
}{
	{nil, []string{}},
	{&When{}, []string{}},

	// Equal
	{
		&When{Equal: eqMap{"foo": sl{"true"}}},
		[]string{"foo"},
	},
	{
		&When{Equal: eqMap{"foo": sl{"true"}, "bar": sl{"true"}}},
		[]string{"foo", "bar"},
	},

	// NotEqual
	{
		&When{NotEqual: eqMap{"foo": sl{"true"}}},
		[]string{"foo"},
	},
	{
		&When{NotEqual: eqMap{"foo": sl{"true"}, "bar": sl{"true"}}},
		[]string{"foo", "bar"},
	},

	// Both
	{
		&When{
			Equal:    eqMap{"foo": sl{"true"}},
			NotEqual: eqMap{"foo": sl{"true"}},
		},
		[]string{"foo"},
	},
	{
		&When{
			Equal:    eqMap{"foo": sl{"true"}},
			NotEqual: eqMap{"bar": sl{"true"}},
		},
		[]string{"foo", "bar"},
	},
}

func TestWhen_Dependencies(t *testing.T) {
	for _, tt := range dependenciestests {
		actual := tt.when.Dependencies()
		if !equalUnordered(tt.expected, actual) {
			t.Errorf(
				"%+v.Dependencies(): expected %s, actual %s",
				tt.when, tt.expected, actual,
			)
		}
	}
}

// nolint: dupl
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

var validatetests = []struct {
	when      *When
	options   map[string]string
	shouldErr bool
}{
	// Empty
	{nil, nil, false},
	{&When{}, nil, false},

	// Command Clauses
	{&When{Command: sl{"test 1 = 1"}}, nil, false},
	{&When{Command: sl{"test 1 = 0"}}, nil, true},
	{&When{Command: sl{"test 1 = 1", "test 0 = 0"}}, nil, false},
	{&When{Command: sl{"test 1 = 1", "test 1 = 0"}}, nil, true},

	// Exist Clauses
	{&When{Exists: sl{"when_test.go"}}, nil, false},
	{&When{Exists: sl{"fakefile"}}, nil, true},
	{&When{Exists: sl{"when_test.go", "fakefile"}}, nil, true},

	// OS Clauses
	{&When{OS: sl{runtime.GOOS}}, nil, false},
	{&When{OS: sl{"fake"}}, nil, true},
	{&When{OS: sl{runtime.GOOS, "fake"}}, nil, false},
	{&When{OS: sl{"fake", runtime.GOOS}}, nil, false},

	// Equal Clauses
	{
		&When{Equal: eqMap{"foo": sl{"true"}}},
		map[string]string{"foo": "true"},
		false,
	},
	{
		&When{Equal: eqMap{"foo": sl{"true"}, "bar": sl{"false"}}},
		map[string]string{"foo": "true", "bar": "false"},
		false,
	},
	{
		&When{Equal: eqMap{"foo": sl{"true"}}},
		map[string]string{"foo": "false"},
		true,
	},
	{
		&When{Equal: eqMap{"foo": sl{"true"}}},
		map[string]string{},
		true,
	},
	{
		&When{Equal: eqMap{"foo": sl{"true"}}},
		map[string]string{"bar": "true"},
		true,
	},
	{
		&When{Equal: eqMap{"foo": sl{"true"}, "bar": sl{"true"}}},
		map[string]string{"bar": "true"},
		true,
	},

	// NotEqual Clauses
	{
		&When{NotEqual: eqMap{"foo": sl{"true"}}},
		map[string]string{"foo": "true"},
		true,
	},
	{
		&When{NotEqual: eqMap{"foo": sl{"true"}}},
		map[string]string{"foo": "false"},
		false,
	},
	{
		&When{NotEqual: eqMap{"foo": sl{"true"}}},
		map[string]string{},
		true,
	},
	{
		&When{NotEqual: eqMap{"foo": sl{"true"}}},
		map[string]string{"bar": "true"},
		true,
	},
	{
		&When{NotEqual: eqMap{"foo": sl{"true"}, "bar": sl{"true"}}},
		map[string]string{"bar": "true"},
		true,
	},
	{
		&When{NotEqual: eqMap{"foo": sl{"true"}, "bar": sl{"true"}}},
		map[string]string{"foo": "false", "bar": "false"},
		false,
	},
}

func TestWhen_Validate(t *testing.T) {
	for _, tt := range validatetests {
		err := tt.when.Validate(tt.options)
		didErr := err != nil
		if tt.shouldErr != didErr {
			t.Errorf(
				"%+v.Validate(): expected error: %t, got error: '%s'",
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
