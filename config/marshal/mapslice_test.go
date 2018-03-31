package marshal

import (
	"errors"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestParseOrderedMap(t *testing.T) {
	index := 0

	ms := yaml.MapSlice{
		{Key: "foo", Value: "bar"},
		{Key: "bar", Value: "baz"},
	}

	defer func() {
		if len(ms) != index {
			t.Errorf("want %d calls to `assign`, got %d", len(ms), index)
		}
	}()

	assign := func(name string, text []byte) error {
		if key, ok := ms[index].Key.(string); !ok || key != name {
			t.Errorf(
				"want key at index %d to be %q, got %q",
				index, key, name,
			)
		}

		if value, ok := ms[index].Value.(string); !ok || value+"\n" != string(text) {
			t.Errorf(
				"want value at index %d to be %q, got %q",
				index, string(value), string(text),
			)
		}

		index++
		return nil
	}

	ParseOrderedMap(ms, assign)
}

func TestParseOrderedMap_stops_on_failure(t *testing.T) {
	index := 0

	ms := yaml.MapSlice{
		{Key: "foo", Value: "bar"},
		{Key: "bar", Value: "baz"},
	}

	defer func() {
		if 1 != index {
			t.Errorf("want 1 call to `assign`, got %d", index)
		}
	}()

	assign := func(name string, text []byte) error {
		index++
		return errors.New("uh oh")
	}

	ParseOrderedMap(ms, assign)
}

func TestParseOrderedMap_validates_key(t *testing.T) {
	ms := yaml.MapSlice{
		{Key: []string{"foo", "bar"}, Value: "bar"},
	}

	assign := func(name string, text []byte) error {
		return nil
	}

	ParseOrderedMap(ms, assign)
}
