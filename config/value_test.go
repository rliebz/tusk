package config

import (
	"reflect"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestValue_UnmarshalYAML(t *testing.T) {
	s1 := []byte(`value: example`)
	s2 := []byte(`example`)
	v1 := Value{}
	v2 := Value{}

	if err := yaml.UnmarshalStrict(s1, &v1); err != nil {
		t.Fatalf("yaml.UnmarshalStrict(%s, ...): unexpected error: %s", s1, err)
	}

	if err := yaml.UnmarshalStrict(s2, &v2); err != nil {
		t.Fatalf("yaml.UnmarshalStrict(%s, ...): unexpected error: %s", s2, err)
	}

	if !reflect.DeepEqual(v1, v2) {
		t.Errorf(
			"Unmarshaling of values `%s` and `%s` not equal:\n%#v != %#v",
			s1, s2, v1, v2,
		)
	}

	if v1.Value != "example" {
		t.Errorf(
			"yaml.UnmarshalStrict(%s, ...): expected member `%s`, actual `%s`",
			s1, "example", v1.Command,
		)
	}
}

func TestValue_UnmarshalYAML_value_and_command(t *testing.T) {
	s := []byte(`{value: "example", command: "echo hello"}`)
	v := Value{}

	if err := yaml.UnmarshalStrict(s, &v); err == nil {
		t.Fatalf(
			"yaml.UnmarshalStrict(%s, ...): expected err, actual nil", s,
		)
	}
}

func TestValueList_UnmarshalYAML(t *testing.T) {
	s1 := []byte(`example`)
	s2 := []byte(`[example]`)
	v1 := ValueList{}
	v2 := ValueList{}

	if err := yaml.UnmarshalStrict(s1, &v1); err != nil {
		t.Fatalf("yaml.UnmarshalStrict(%s, ...): unexpcted error: %s", s1, err)
	}

	if err := yaml.UnmarshalStrict(s2, &v2); err != nil {
		t.Fatalf("yaml.UnmarshalStrict(%s, ...): unexpcted error: %s", s2, err)
	}

	if !reflect.DeepEqual(v1, v2) {
		t.Errorf(
			"Unmarshaling of valueLists `%s` and `%s` not equal:\n%#v != %#v",
			s1, s2, v1, v2,
		)
	}

	if len(v1) != 1 {
		t.Errorf(
			"yaml.UnmarshalStrict(%s, ...): expected 1 item, actual %d",
			s1, len(v1),
		)
	}

	if v1[0].Value != "example" {
		t.Errorf(
			"yaml.UnmarshalStrict(%s, ...): expected member `%s`, actual `%s`",
			s1, "example", v1[0].Value,
		)
	}
}
