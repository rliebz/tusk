package task

import (
	"reflect"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestSubTask_UnmarshalYAML(t *testing.T) {
	s1 := []byte(`name: example`)
	s2 := []byte(`example`)
	st1 := SubTask{}
	st2 := SubTask{}

	if err := yaml.Unmarshal(s1, &st1); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s1, err)
	}

	if err := yaml.Unmarshal(s2, &st2); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s2, err)
	}

	if !reflect.DeepEqual(st1, st2) {
		t.Errorf(
			"Unmarshalling of subtasks `%s` and `%s` not equal:\n%#v != %#v",
			s1, s2, st1, st2,
		)
	}

	if st1.Name != "example" {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected member `%s`, actual `%s`",
			s1, "example", st1.Name,
		)
	}
}

func TestSubTaskList_UnmarshalYAML(t *testing.T) {
	s1 := []byte(`example`)
	s2 := []byte(`[example]`)
	l1 := SubTaskList{}
	l2 := SubTaskList{}

	if err := yaml.Unmarshal(s1, &l1); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s1, err)
	}

	if err := yaml.Unmarshal(s2, &l2); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s2, err)
	}

	if !reflect.DeepEqual(l1, l2) {
		t.Errorf(
			"Unmarshalling of SubTaskLists `%s` and `%s` not equal:\n%#v != %#v",
			s1, s2, l1, l2,
		)
	}

	if len(l1) != 1 {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected 1 item, actual %d",
			s1, len(l1),
		)
	}

	if l1[0].Name != "example" {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected member `%s`, actual `%v`",
			s1, "example", l1[0].Name,
		)
	}
}
