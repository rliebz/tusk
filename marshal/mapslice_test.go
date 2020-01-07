package marshal

import (
	"errors"
	"reflect"
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
				index, value, string(text),
			)
		}

		index++
		return nil
	}

	actual, err := ParseOrderedMap(ms, assign)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	expected := []string{"foo", "bar"}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("want %v, got %v", expected, actual)
	}
}

func TestParseOrderedMap_stops_on_failure(t *testing.T) {
	index := 0

	ms := yaml.MapSlice{
		{Key: "foo", Value: "bar"},
		{Key: "bar", Value: "baz"},
	}

	defer func() {
		if index != 1 {
			t.Errorf("want 1 call to `assign`, got %d", index)
		}
	}()

	assign := func(name string, text []byte) error {
		index++
		return errors.New("uh oh")
	}

	if _, err := ParseOrderedMap(ms, assign); err == nil {
		t.Fatal("want error \"uh oh\", got nil")
	}
}

func TestParseOrderedMap_validates_key(t *testing.T) {
	ms := yaml.MapSlice{
		{Key: []string{"foo", "bar"}, Value: "bar"},
	}

	assign := func(name string, text []byte) error {
		return nil
	}

	if _, err := ParseOrderedMap(ms, assign); err == nil {
		t.Fatal("want error for invalid key, got nil")
	}
}
