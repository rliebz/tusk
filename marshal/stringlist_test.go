package marshal

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	yaml "gopkg.in/yaml.v2"
)

type StringListHolder struct {
	Foo StringList
}

func TestStringList(t *testing.T) {
	g := ghost.New(t)

	var h1 StringListHolder
	err := yaml.UnmarshalStrict([]byte(`foo: example`), &h1)
	g.NoError(err)

	var h2 StringListHolder
	err = yaml.UnmarshalStrict([]byte(`foo: [example]`), &h2)
	g.NoError(err)

	g.Should(be.DeepEqual(h1, h2))
	g.Should(be.DeepEqual(StringListHolder{Foo: StringList{"example"}}, h1))
}

func TestStringList_fails(t *testing.T) {
	g := ghost.New(t)

	var h1 StringListHolder
	err := yaml.UnmarshalStrict([]byte(`foo: [bar: baz]`), &h1)
	g.Should(be.ErrorContaining("cannot unmarshal !!map into string", err))

	var h2 StringListHolder
	err = yaml.UnmarshalStrict([]byte(`foo: bar: baz`), &h2)
	g.Should(be.ErrorContaining("mapping values are not allowed in this context", err))
}

func TestNullableStringList(t *testing.T) {
	g := ghost.New(t)

	var nsl1 NullableStringList
	err := yaml.UnmarshalStrict([]byte(`example`), &nsl1)
	g.NoError(err)

	var nsl2 NullableStringList
	err = yaml.UnmarshalStrict([]byte(`[example]`), &nsl2)
	g.NoError(err)

	g.Should(be.DeepEqual(nsl1, nsl2))

	want := "example"
	g.Should(be.DeepEqual(NullableStringList{&want}, nsl1))
}

func TestNullableStringList_null(t *testing.T) {
	g := ghost.New(t)

	var nsl NullableStringList
	err := yaml.UnmarshalStrict([]byte("~"), &nsl)
	g.NoError(err)

	g.Should(be.SliceLen(0, nsl))
}

func TestNullableStringList_null_item(t *testing.T) {
	g := ghost.New(t)

	var nsl NullableStringList
	err := yaml.UnmarshalStrict([]byte("[one, null, two]"), &nsl)
	g.NoError(err)

	if g.Should(be.SliceLen(3, nsl)) {
		g.Should(be.Nil(nsl[1]))
	}
}

func TestNullableStringList_fails(t *testing.T) {
	g := ghost.New(t)

	var nsl1 NullableStringList
	err := yaml.UnmarshalStrict([]byte(`[bar: baz]`), &nsl1)
	g.Should(be.ErrorContaining("cannot unmarshal !!map into string", err))

	var nsl2 NullableStringList
	err = yaml.UnmarshalStrict([]byte(`foo: bar: baz`), &nsl2)
	g.Should(be.ErrorEqual("yaml: mapping values are not allowed in this context", err))
}
