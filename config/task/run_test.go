package task

import (
	"reflect"
	"testing"

	"github.com/rliebz/tusk/config/when"
	"gopkg.in/yaml.v2"
)

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

var environmentTests = []struct {
	input          string
	expectedLength int
}{
	{`{}`, 0},
	{`{environment: {}}`, 0},
	{`{set_environment: {}}`, 0},
	{`{environment: {foo: bar}}`, 1},
	{`{set_environment: {foo: bar}}`, 1},
	{`{environment: {foo: bar, bar: baz}}`, 2},
	{`{set_environment: {foo: bar, bar: baz}}`, 2},
}

func TestRun_UnmarshalYAML_environment_and_set_environment(t *testing.T) {
	for _, testCase := range environmentTests {
		r := Run{}

		if err := yaml.Unmarshal([]byte(testCase.input), &r); err != nil {
			t.Errorf(
				"yaml.Unmarshal(%s, ...): unexpected error: %s",
				testCase.input, err,
			)
			continue
		}

		if testCase.expectedLength != len(r.SetEnvironment) {
			t.Errorf(
				"yaml.Unmarshal(%s, ...): expected %d environment items, got %d",
				testCase.input, testCase.expectedLength, len(r.SetEnvironment),
			)
		}
	}
}

var multipleActionTests = []string{
	`{command: example, task: echo 'hello'}`,
	`{command: example, environment: {foo: bar}}`,
	`{task: echo 'hello', environment: {foo: bar}}`,
	`{command: example, task: echo 'hello', environment: {foo: bar}}`,
	`{environment: {foo: bar}, set_environment: {bar: baz}}`,
}

func TestRun_UnmarshalYAML_command_and_subtask(t *testing.T) {
	for _, input := range multipleActionTests {
		r := Run{}
		if err := yaml.Unmarshal([]byte(input), &r); err == nil {
			t.Errorf(
				"yaml.Unmarshal(%s, ...): expected error, received nil",
				input,
			)
		}
	}
}

var shouldtests = []struct {
	desc     string
	input    Run
	expected bool
	vars     map[string]string
}{
	{"no when clause", Run{}, true, nil},
	{"true when clause", Run{When: when.List{when.True}}, true, nil},
	{"false when clause", Run{When: when.List{when.False}}, false, nil},
	{
		"var matches condition",
		Run{When: when.List{when.Create(when.WithEqual("foo", "bar"))}},
		true,
		map[string]string{"foo": "bar"},
	},
	{
		"var does not match condition",
		Run{When: when.List{when.Create(when.WithEqual("foo", "bar"))}},
		false,
		map[string]string{"foo": "baz"},
	},
	{
		"var was not passed",
		Run{When: when.List{when.Create(when.WithEqual("foo", "bar"))}},
		false,
		nil,
	},
}

func TestRun_shouldRun(t *testing.T) {
	for _, tt := range shouldtests {
		actual, err := tt.input.shouldRun(tt.vars)
		if err != nil {
			t.Errorf(
				"task.shouldRun() for %s: unexpected error: %s",
				tt.desc, err,
			)
			continue
		}
		if tt.expected != actual {
			t.Errorf(
				"task.shouldRun() for %s: expected: %t, actual: %t",
				tt.desc, tt.expected, actual,
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
