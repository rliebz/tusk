package appyaml

import (
	"runtime"
	"testing"
)

var whentests = []struct {
	when      When
	shouldErr bool
}{
	// Empty
	{When{}, false},

	// Command Clauses
	{When{Command: StringList{[]string{"test 1 = 1"}}}, false},
	{When{Command: StringList{[]string{"test 1 = 0"}}}, true},
	{When{Command: StringList{[]string{"test 1 = 1", "test 0 = 0"}}}, false},
	{When{Command: StringList{[]string{"test 1 = 1", "test 1 = 0"}}}, true},

	// Exist Clauses
	{When{Exists: StringList{[]string{"when_test.go"}}}, false},
	{When{Exists: StringList{[]string{"fakefile"}}}, true},
	{When{Exists: StringList{[]string{"when_test.go", "fakefile"}}}, true},

	// OS Clauses
	{When{OS: StringList{[]string{runtime.GOOS}}}, false},
	{When{OS: StringList{[]string{"fake"}}}, true},
	{When{OS: StringList{[]string{runtime.GOOS, "fake"}}}, false},
	{When{OS: StringList{[]string{"fake", runtime.GOOS}}}, false},
}

func TestWhen_Validate(t *testing.T) {
	for _, tt := range whentests {
		err := tt.when.Validate()
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
