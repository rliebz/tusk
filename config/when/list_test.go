package when

import (
	"reflect"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestList_UnmarshalYAML(t *testing.T) {
	s1 := []byte(`os: linux`)
	s2 := []byte(`[os: linux]`)
	l1 := List{}
	l2 := List{}

	if err := yaml.Unmarshal(s1, &l1); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s1, err)
	}

	if err := yaml.Unmarshal(s2, &l2); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s2, err)
	}

	if !reflect.DeepEqual(l1, l2) {
		t.Errorf(
			"Unmarshalling of Lists `%s` and `%s` not equal:\n%#v != %#v",
			s1, s2, l1, l2,
		)
	}

	if len(l1) != 1 {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected 1 item, actual %d",
			s1, len(l1),
		)
	}

	if l1[0].OS[0] != "linux" {
		t.Errorf(
			"yaml.Unmarshal(%s, ...): expected os `%s`, actual `%v`",
			s1, "linux", l1[0].OS[0],
		)
	}
}

var listDepTests = []struct {
	testCase string
	list     List
	expected []string
}{
	{
		"empty list",
		List{},
		[]string{},
	},
	{
		"single item list",
		List{
			Create(WithEqual("foo", "true"), WithEqual("bar", "true")),
		},
		[]string{"foo", "bar"},
	},
	{
		"duplicate across lists",
		List{
			Create(WithEqual("foo", "true")),
			Create(WithEqual("foo", "true")),
		},
		[]string{"foo"},
	},
	{
		"different items per list",
		List{
			Create(WithEqual("foo", "true")),
			Create(WithEqual("bar", "true")),
		},
		[]string{"foo", "bar"},
	},
}

func TestList_Dependencies(t *testing.T) {
	for _, tt := range listDepTests {
		actual := tt.list.Dependencies()
		if !equalUnordered(tt.expected, actual) {
			t.Errorf(
				"List.Dependencies() for %s: expected %s, actual %s",
				tt.testCase, tt.expected, actual,
			)
		}
	}
}

func TestList_Dependencies_nil(t *testing.T) {
	var l *List
	actual := l.Dependencies()
	if len(actual) > 0 {
		t.Errorf("expected 0 dependencies, got: %s", actual)
	}
}

var listValidateTests = []struct {
	testCase  string
	list      List
	options   map[string]string
	shouldErr bool
}{
	{
		"all valid",
		List{True, True, True},
		nil,
		false,
	},
	{
		"all invalid",
		List{False, False, False},
		nil,
		true,
	},
	{
		"some invalid",
		List{True, False, True},
		nil,
		true,
	},
	{
		"passes requirements",
		List{
			Create(WithEqual("foo", "true")),
			Create(WithEqual("bar", "false")),
		},
		map[string]string{"foo": "true", "bar": "false"},
		false,
	},
}

func TestList_Validate(t *testing.T) {
	for _, tt := range listValidateTests {
		err := tt.list.Validate(tt.options)
		didErr := err != nil
		if tt.shouldErr != didErr {
			t.Errorf(
				"list.Validate() for %s: expected error: %t, got error: '%s'",
				tt.testCase, tt.shouldErr, err,
			)
		}
	}
}

func TestList_Validate_nil(t *testing.T) {
	var l *List
	if err := l.Validate(map[string]string{}); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
