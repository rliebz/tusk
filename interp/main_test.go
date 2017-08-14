package interp

import (
	"bytes"
	"testing"
)

var maptests = []struct {
	input    []byte
	vars     map[string]string
	expected []byte
}{
	{
		[]byte("${foo}"),
		map[string]string{"foo": "bar"},
		[]byte("bar"),
	},
	{
		[]byte("foo"),
		map[string]string{"foo": "bar"},
		[]byte("foo"),
	},
	{
		[]byte("$foo"),
		map[string]string{"foo": "bar"},
		[]byte("$foo"),
	},
	{
		[]byte("${foo}${foo}"),
		map[string]string{"foo": "bar"},
		[]byte("barbar"),
	},
	{
		[]byte("${foo}${bar}"),
		map[string]string{"foo": "bar"},
		[]byte("bar${bar}"),
	},
	{
		[]byte("$${foo}"),
		map[string]string{"foo": "bar"},
		[]byte("${foo}"),
	},

	{
		[]byte("$$${foo}"),
		map[string]string{"foo": "bar"},
		[]byte("$bar"),
	},
	{
		[]byte("$"),
		map[string]string{"foo": "bar"},
		[]byte("$"),
	},
	{
		[]byte("$$"),
		map[string]string{"foo": "bar"},
		[]byte("$"),
	},
}

func TestMap(t *testing.T) {
	for _, tt := range maptests {
		actual, err := Map(tt.input, tt.vars)
		if err != nil {
			t.Errorf("Unexpected err: %e", err)
		}

		if !bytes.Equal(tt.expected, actual) {
			t.Errorf(
				"Map(%s): expected: %s, actual: %s",
				string(tt.input), string(tt.expected), string(actual),
			)
		}
	}
}

var containstests = []struct {
	input    []byte
	name     string
	expected bool
}{
	{[]byte("${foo}"), "foo", true},
	{[]byte("${bar}"), "foo", false},
	{[]byte("foo"), "foo", false},
	{[]byte("$${foo}"), "foo", false},
	{[]byte("$foo"), "foo", false},
}

func TestContains(t *testing.T) {
	for _, tt := range containstests {
		actual, err := Contains(tt.input, tt.name)
		if err != nil {
			t.Errorf("Unexpected err: %e", err)
		}

		if tt.expected != actual {
			t.Errorf(
				"Contains(%s, %s): expected: %t, actual: %t",
				string(tt.input), tt.name, tt.expected, actual,
			)
		}
	}
}
