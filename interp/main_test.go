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
		[]byte("${foo}${foo}"),
		map[string]string{"foo": "bar"},
		[]byte("barbar"),
	},
	{
		[]byte("${foo}${bar}"),
		map[string]string{"foo": "bar"},
		[]byte("bar${bar}"),
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
