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

	// Exist Clauses
	{When{Exists: []string{"when_test.go"}}, false},
	{When{Exists: []string{"fakefile"}}, true},
	{When{Exists: []string{"when_test.go", "fakefile"}}, true},

	// OS Clauses
	{When{OS: []string{runtime.GOOS}}, false},
	{When{OS: []string{"fake"}}, true},
	{When{OS: []string{runtime.GOOS, "fake"}}, false},
	{When{OS: []string{"fake", runtime.GOOS}}, false},

	// Test Clauses
	{When{Test: []string{"1 = 1"}}, false},
	{When{Test: []string{"1 = 0"}}, true},
	{When{Test: []string{"1 = 1", "0 = 0"}}, false},
	{When{Test: []string{"1 = 1", "1 = 0"}}, true},
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
