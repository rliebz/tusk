package marshal

import (
	"errors"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	yaml "gopkg.in/yaml.v2"
)

func TestParseOrderedMap(t *testing.T) {
	g := ghost.New(t)

	index := 0
	ms := yaml.MapSlice{
		{Key: "foo", Value: "bar"},
		{Key: "bar", Value: "baz"},
	}

	defer func() { g.Should(be.Equal(2, index)) }()

	assign := func(name string, text []byte) error {
		defer func() { index++ }()

		key, ok := ms[index].Key.(string)
		g.Check(ok)

		g.Should(be.Equal(key, name))

		value, ok := ms[index].Value.(string)
		g.Check(ok)

		g.Should(be.Equal(string(text), value+"\n"))

		return nil
	}

	got, err := ParseOrderedMap(ms, assign)
	g.NoError(err)

	want := []string{"foo", "bar"}
	g.Should(be.DeepEqual(got, want))
}

func TestParseOrderedMap_stops_on_failure(t *testing.T) {
	g := ghost.New(t)

	index := 0
	ms := yaml.MapSlice{
		{Key: "foo", Value: "bar"},
		{Key: "bar", Value: "baz"},
	}

	defer func() { g.Should(be.Equal(1, index)) }()

	assign := func(string, []byte) error {
		index++
		return errors.New("uh oh")
	}

	_, err := ParseOrderedMap(ms, assign)
	g.Should(be.ErrorEqual(err, "uh oh"))
}

func TestParseOrderedMap_validates_key(t *testing.T) {
	g := ghost.New(t)

	ms := yaml.MapSlice{
		{Key: []string{"foo", "bar"}, Value: "bar"},
	}

	assign := func(string, []byte) error {
		return nil
	}

	_, err := ParseOrderedMap(ms, assign)
	g.Should(be.ErrorEqual(err, `["foo" "bar"] is not a valid key name`))
}
