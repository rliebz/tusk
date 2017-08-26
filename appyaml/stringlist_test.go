package appyaml

import "testing"
import "gopkg.in/yaml.v2"
import "reflect"

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
			"Unmarshalling of StringLists `%s` and `%s` not equal:\n%#v != %#v",
			s1, s2, h1, h2,
		)
	}

	if len(h1.Foo.Values) != 1 {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected 1 item, actual %d",
			s1, len(h1.Foo.Values),
		)
	}

	if h1.Foo.Values[0] != "example" {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected member `%s`, actual `%s`",
			s1, "example", h1.Foo.Values[0],
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
