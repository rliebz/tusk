package marshal

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	yaml "gopkg.in/yaml.v2"
)

type SliceHolder[T any] struct {
	Foo Slice[T]
}

func TestSlice(t *testing.T) {
	g := ghost.New(t)

	var h1 SliceHolder[string]
	err := yaml.UnmarshalStrict([]byte(`foo: example`), &h1)
	g.NoError(err)

	var h2 SliceHolder[string]
	err = yaml.UnmarshalStrict([]byte(`foo: [example]`), &h2)
	g.NoError(err)

	g.Should(be.DeepEqual(h1, h2))
	g.Should(be.DeepEqual(h1, SliceHolder[string]{Foo: Slice[string]{"example"}}))
}

func TestSlice_error(t *testing.T) {
	g := ghost.New(t)

	var h1 SliceHolder[string]
	err := yaml.UnmarshalStrict([]byte(`foo: [bar: baz]`), &h1)
	g.Should(be.ErrorContaining(err, "cannot unmarshal !!seq into string"))

	var h2 SliceHolder[string]
	err = yaml.UnmarshalStrict([]byte(`foo: bar: baz`), &h2)
	g.Should(be.ErrorContaining(err, "mapping values are not allowed in this context"))
}

func TestSlice_pointer(t *testing.T) {
	g := ghost.New(t)

	var nsl1 Slice[*string]
	err := yaml.UnmarshalStrict([]byte(`example`), &nsl1)
	g.NoError(err)

	var nsl2 Slice[*string]
	err = yaml.UnmarshalStrict([]byte(`[example]`), &nsl2)
	g.NoError(err)

	g.Should(be.DeepEqual(nsl1, nsl2))

	want := "example"
	g.Should(be.DeepEqual(nsl1, Slice[*string]{&want}))
}

func TestSlice_pointer_null(t *testing.T) {
	g := ghost.New(t)

	var nsl Slice[*string]
	err := yaml.UnmarshalStrict([]byte("~"), &nsl)
	g.NoError(err)

	g.Should(be.SliceLen(nsl, 0))
}

func TestSlice_pointer_single_null_item(t *testing.T) {
	g := ghost.New(t)

	var nsl Slice[*string]
	err := yaml.UnmarshalStrict([]byte("[null]"), &nsl)
	g.NoError(err)

	if g.Should(be.SliceLen(nsl, 1)) {
		g.Should(be.Nil(nsl[0]))
	}
}

func TestSlice_pointer_null_item(t *testing.T) {
	g := ghost.New(t)

	var nsl Slice[*string]
	err := yaml.UnmarshalStrict([]byte("[one, null, two]"), &nsl)
	g.NoError(err)

	if g.Should(be.SliceLen(nsl, 3)) {
		g.Should(be.Nil(nsl[1]))
	}
}

func TestSlice_pointer_error(t *testing.T) {
	g := ghost.New(t)

	var nsl1 Slice[*string]
	err := yaml.UnmarshalStrict([]byte(`[bar: baz]`), &nsl1)
	g.Should(be.ErrorContaining(err, "cannot unmarshal !!seq into string"))

	var nsl2 Slice[*string]
	err = yaml.UnmarshalStrict([]byte(`foo: bar: baz`), &nsl2)
	g.Should(be.ErrorEqual(err, "yaml: mapping values are not allowed in this context"))
}
