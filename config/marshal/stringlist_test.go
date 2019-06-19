package marshal

import (
	"reflect"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

type StringListHolder struct {
	Foo StringList
}

func TestStringList(t *testing.T) {
	s1 := []byte(`foo: example`)
	s2 := []byte(`foo: [example]`)

	h1 := StringListHolder{}
	h2 := StringListHolder{}

	if err := yaml.Unmarshal(s1, &h1); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpcted error: %s", s1, err)
	}

	if err := yaml.Unmarshal(s2, &h2); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpcted error: %s", s2, err)
	}

	if !reflect.DeepEqual(h1, h2) {
		t.Errorf(
			"Unmarshaling of StringLists `%s` and `%s` not equal:\n%#v != %#v",
			s1, s2, h1, h2,
		)
	}

	if len(h1.Foo) != 1 {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected 1 item, actual %d",
			s1, len(h1.Foo),
		)
	}

	if h1.Foo[0] != "example" {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected member `%s`, actual `%s`",
			s1, "example", h1.Foo[0],
		)
	}
}

func TestStringList_fails(t *testing.T) {
	s1 := []byte(`foo: [bar: baz]`)
	h1 := StringListHolder{}

	if err := yaml.Unmarshal(s1, &h1); err == nil {
		t.Errorf("yaml.Unmarshal(%s, ...): expected error, got nil", s1)
	}

	s2 := []byte(`foo: bar: baz`)
	h2 := StringListHolder{}

	if err := yaml.Unmarshal(s2, &h2); err == nil {
		t.Errorf("yaml.Unmarshal(%s, ...): expected error, got nil", s2)
	}
}

func TestNullableStringList(t *testing.T) {
	s1 := []byte(`example`)
	s2 := []byte(`[example]`)

	nsl1 := NullableStringList{}
	nsl2 := NullableStringList{}

	if err := yaml.Unmarshal(s1, &nsl1); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s1, err)
	}

	if err := yaml.Unmarshal(s2, &nsl2); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s2, err)
	}

	if !reflect.DeepEqual(nsl1, nsl2) {
		t.Errorf(
			"Unmarshaling of NullableStringLists `%s` and `%s` not equal:\n%#v != %#v",
			s1, s2, nsl1, nsl2,
		)
	}

	if len(nsl1) != 1 {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected 1 item, actual %d",
			s1, len(nsl1),
		)
	}

	if *nsl1[0] != "example" {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected member `%s`, actual `%s`",
			s1, "example", *nsl1[0],
		)
	}
}

func TestNullableStringList_null(t *testing.T) {
	s := []byte("~")

	nsl := NullableStringList{}

	if err := yaml.Unmarshal(s, &nsl); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s, err)
	}

	if len(nsl) != 0 {
		t.Fatalf("expected 0 items, got %d", len(nsl))
	}
}

func TestNullableStringList_null_item(t *testing.T) {
	s := []byte("[one, null, two]")
	nsl := NullableStringList{}

	if err := yaml.Unmarshal(s, &nsl); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s, err)
	}

	if len(nsl) != 3 {
		t.Fatalf("expected 3 items, got %d", len(nsl))
	}

	if nsl[1] != nil {
		t.Errorf("expected nil, got %s", *nsl[1])
	}
}

func TestNullableStringList_fails(t *testing.T) {
	s1 := []byte(`[bar: baz]`)
	nsl1 := NullableStringList{}

	if err := yaml.Unmarshal(s1, &nsl1); err == nil {
		t.Errorf("yaml.Unmarshal(%s, ...): expected error, got nil", s1)
	}

	s2 := []byte(`foo: bar: baz`)
	nsl2 := NullableStringList{}

	if err := yaml.Unmarshal(s2, &nsl2); err == nil {
		t.Errorf("yaml.Unmarshal(%s, ...): expected error, got nil", s2)
	}
}
