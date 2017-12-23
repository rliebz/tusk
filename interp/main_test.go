package interp

import (
	"bytes"
	"reflect"
	"testing"
)

var escapetests = []struct {
	input    string
	expected string
}{
	{"$", "$"},
	{"$$", "$"},
	{"$$$", "$$"},
}

func TestEscape(t *testing.T) {
	for _, tt := range escapetests {
		escaped := escape([]byte(tt.input))
		actual := string(escaped)

		if tt.expected != actual {
			t.Errorf(
				"Escape(%s): expected: %s, actual: %s",
				tt.input, tt.expected, actual,
			)
		}
	}
}

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
		[]byte("$${foo}"),
	},

	{
		[]byte("$$${foo}"),
		map[string]string{"foo": "bar"},
		[]byte("$$bar"),
	},
	{
		[]byte("$"),
		map[string]string{"foo": "bar"},
		[]byte("$"),
	},
	{
		[]byte("$$"),
		map[string]string{"foo": "bar"},
		[]byte("$$"),
	},
}

func TestMap(t *testing.T) {
	for _, tt := range maptests {
		actual, err := mapInterpolate(tt.input, tt.vars)
		if err != nil {
			t.Errorf("Unexpected err: %s", err)
			continue
		}

		if !bytes.Equal(tt.expected, actual) {
			t.Errorf(
				"Map(%s): expected: %s, actual: %s",
				string(tt.input), string(tt.expected), string(actual),
			)
		}
	}
}

var findtests = []struct {
	input    []byte
	expected []string
}{
	{[]byte(""), []string{}},
	{[]byte("${}"), []string{}},
	{[]byte("foo"), []string{}},
	{[]byte("$foo"), []string{}},
	{[]byte("${foo}"), []string{"foo"}},
	{[]byte("${f-o-o}"), []string{"f-o-o"}},
	{[]byte("${f_o_o}"), []string{"f_o_o"}},
	{[]byte("${foo}${bar}"), []string{"foo", "bar"}},
	{[]byte("${foo}${FOO}"), []string{"foo", "FOO"}},
	{[]byte("_-${foo}.  ${bar} baz"), []string{"foo", "bar"}},
}

func TestFindPotentialVariables(t *testing.T) {
	for _, tt := range findtests {
		actual := FindPotentialVariables(tt.input)
		if !reflect.DeepEqual(tt.expected, actual) {
			t.Errorf(
				`FindPotentialVariables("%s"): expected: %v, actual %v`,
				string(tt.input), tt.expected, actual,
			)
		}
	}
}
