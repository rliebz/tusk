package when

import (
	"fmt"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestList_UnmarshalYAML(t *testing.T) {
	var unmarshalTests = []struct {
		desc     string
		input    string
		expected List
	}{
		{
			"single item",
			"os: linux",
			List{Create(WithOS("linux"))},
		},
		{
			"list length 1",
			"[os: linux]",
			List{Create(WithOS("linux"))},
		},
		{
			"single item short",
			"foo",
			List{Create(WithEqual("foo", "true"))},
		},
		{
			"list implies multiple whens",
			"[foo, bar]",
			List{Create(WithEqual("foo", "true")), Create(WithEqual("bar", "true"))},
		},
		{
			"nested short lists",
			"[[foo, bar], [baz]]",
			List{
				Create(WithEqual("foo", "true"), WithEqual("bar", "true")),
				Create(WithEqual("baz", "true")),
			},
		},
	}

	for _, tt := range unmarshalTests {
		t.Run(tt.desc, func(t *testing.T) {
			l := List{}
			if err := yaml.Unmarshal([]byte(tt.input), &l); err != nil {
				t.Fatalf(
					`Unmarshalling %s: unexpected error: %s`,
					tt.desc, err,
				)
			}

			// Rely on string representation of When for comparison
			expected := fmt.Sprintf("%s", tt.expected)
			actual := fmt.Sprintf("%s", l)

			if expected != actual {
				t.Errorf("want %q, got %q", expected, actual)
			}
		})
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
