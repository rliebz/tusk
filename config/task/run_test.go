package task

import "testing"
import "gopkg.in/yaml.v2"
import "reflect"

func TestRun_UnmarshalYAML(t *testing.T) {
	s1 := []byte(`command: example`)
	s2 := []byte(`example`)
	r1 := Run{}
	r2 := Run{}

	if err := yaml.Unmarshal(s1, &r1); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s1, err)
	}

	if err := yaml.Unmarshal(s2, &r2); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s2, err)
	}

	if !reflect.DeepEqual(r1, r2) {
		t.Errorf(
			"Unmarshalling of runs `%s` and `%s` not equal:\n%#v != %#v",
			s1, s2, r1, r2,
		)
	}

	if len(r1.Command) != 1 {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected 1 item, actual %d",
			s1, len(r1.Command),
		)
	}

	if r1.Command[0] != "example" {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected member `%s`, actual `%s`",
			s1, "example", r1.Command,
		)
	}
}

var multipleActionTests = []string{
	`{command: example, task: echo 'hello'}`,
	`{command: example, environment: {foo: bar}}`,
	`{task: echo 'hello', environment: {foo: bar}}`,
	`{command: example, task: echo 'hello', environment: {foo: bar}}`,
}

func TestRun_UnmarshalYAML_command_and_subtask(t *testing.T) {
	r := Run{}

	for _, input := range multipleActionTests {
		if err := yaml.Unmarshal([]byte(input), &r); err == nil {
			t.Fatalf(
				"yaml.Unmarshal(%s, ...): expected error, received nil",
				input,
			)
		}
	}
}

type runListHolder struct {
	Foo RunList
}

func TestRunList_UnmarshalYAML(t *testing.T) {
	s1 := []byte(`foo: example`)
	s2 := []byte(`foo: [example]`)

	h1 := runListHolder{}
	h2 := runListHolder{}

	if err := yaml.Unmarshal(s1, &h1); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s1, err)
	}

	if err := yaml.Unmarshal(s2, &h2); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s2, err)
	}

	if !reflect.DeepEqual(h1, h2) {
		t.Errorf(
			"Unmarshalling of runLists `%s` and `%s` not equal:\n%#v != %#v",
			s1, s2, h1, h2,
		)
	}

	if len(h1.Foo) != 1 {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected 1 item, actual %d",
			s1, len(h1.Foo),
		)
	}

	if len(h1.Foo[0].Command) != 1 {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected 1 command, actual %d",
			s1, len(h1.Foo[0].Command),
		)
	}

	if h1.Foo[0].Command[0] != "example" {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected member `%s`, actual `%v`",
			s1, "example", h1.Foo[0],
		)
	}
}
