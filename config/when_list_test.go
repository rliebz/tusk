package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	yaml "gopkg.in/yaml.v2"
)

func TestList_UnmarshalYAML(t *testing.T) {
	unmarshalTests := []struct {
		desc     string
		input    string
		expected List
	}{
		{
			"single item",
			"os: linux",
			List{createWhen(withWhenOS("linux"))},
		},
		{
			"list length 1",
			"[os: linux]",
			List{createWhen(withWhenOS("linux"))},
		},
		{
			"single item short",
			"foo",
			List{createWhen(withWhenEqual("foo", "true"))},
		},
		{
			"list implies multiple whens",
			"[foo, bar]",
			List{createWhen(withWhenEqual("foo", "true")), createWhen(withWhenEqual("bar", "true"))},
		},
		{
			"nested short lists",
			"[[foo, bar], [baz]]",
			List{
				createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "true")),
				createWhen(withWhenEqual("baz", "true")),
			},
		},
	}

	for _, tt := range unmarshalTests {
		t.Run(tt.desc, func(t *testing.T) {
			l := List{}
			if err := yaml.UnmarshalStrict([]byte(tt.input), &l); err != nil {
				t.Fatalf(
					`Unmarshaling %s: unexpected error: %s`,
					tt.desc, err,
				)
			}

			if !cmp.Equal(l, tt.expected) {
				t.Errorf("unmarshal mismatch:\n%s", cmp.Diff(tt.expected, l))
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
			createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "true")),
		},
		[]string{"foo", "bar"},
	},
	{
		"duplicate across lists",
		List{
			createWhen(withWhenEqual("foo", "true")),
			createWhen(withWhenEqual("foo", "true")),
		},
		[]string{"foo"},
	},
	{
		"different items per list",
		List{
			createWhen(withWhenEqual("foo", "true")),
			createWhen(withWhenEqual("bar", "true")),
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
		List{whenTrue, whenTrue, whenTrue},
		nil,
		false,
	},
	{
		"all invalid",
		List{whenFalse, whenFalse, whenFalse},
		nil,
		true,
	},
	{
		"some invalid",
		List{whenTrue, whenFalse, whenTrue},
		nil,
		true,
	},
	{
		"passes requirements",
		List{
			createWhen(withWhenEqual("foo", "true")),
			createWhen(withWhenEqual("bar", "false")),
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
