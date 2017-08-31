package task

import (
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/rliebz/tusk/appyaml"
)

func TestOption_Dependencies(t *testing.T) {
	option := &Option{Computed: []struct {
		When    appyaml.When
		content `yaml:",inline"`
	}{
		{When: falseWhen, content: content{Default: "foo"}},
		{When: appyaml.When{
			Equal: map[string]appyaml.StringList{
				"foo": {Values: []string{"foovalue"}},
				"bar": {Values: []string{"barvalue"}},
			},
		}, content: content{Default: "bar"}},
		{When: appyaml.When{
			NotEqual: map[string]appyaml.StringList{
				"baz": {Values: []string{"bazvalue"}},
			},
		}, content: content{Default: "bar"}},
	}}

	expected := []string{"foo", "bar", "baz"}
	actual := option.Dependencies()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf(
			"Option.Dependencies(): expected %s, actual %s",
			expected, actual,
		)
	}
}

// TODO: Make these more accessible to other tests
var trueWhen = appyaml.When{OS: appyaml.StringList{Values: []string{runtime.GOOS}}}
var falseWhen = appyaml.When{OS: appyaml.StringList{Values: []string{"FAKE"}}}

// Env var `OPTION_VAR` will be set to `option_val`
var valuetests = []struct {
	desc     string
	input    *Option
	expected string
}{
	{"nil", nil, ""},
	{"empty option", &Option{}, ""},
	{
		"default only",
		&Option{content: content{Default: "default"}},
		"default",
	},
	{
		"command only",
		&Option{content: content{Command: "echo command"}},
		"command",
	},
	{
		"environment variable only",
		&Option{Environment: "OPTION_VAR"},
		"option_val",
	},
	{
		"passed variable only",
		&Option{Passed: "passed"},
		"passed",
	},
	{
		"computed value",
		&Option{Computed: []struct {
			When    appyaml.When
			content `yaml:",inline"`
		}{
			{When: falseWhen, content: content{Default: "foo"}},
			{When: trueWhen, content: content{Default: "bar"}},
			{When: falseWhen, content: content{Default: "baz"}},
		}},
		"bar",
	},
	{
		"computed fallthrough to default",
		&Option{content: content{Default: "default"}, Computed: []struct {
			When    appyaml.When
			content `yaml:",inline"`
		}{
			{When: falseWhen, content: content{Default: "false when"}},
		}},
		"default",
	},
	{
		"passed when all settings are defined",
		&Option{
			content:     content{Default: "default"},
			Environment: "OPTION_VAR",
			Computed: []struct {
				When    appyaml.When
				content `yaml:",inline"`
			}{
				{When: trueWhen, content: content{Default: "when"}},
			},
			Passed: "passed",
		},
		"passed",
	},
}

func TestOption_Value(t *testing.T) {
	if err := os.Setenv("OPTION_VAR", "option_val"); err != nil {
		t.Fatalf("unexpected err setting environment variable: %s", err)
	}

	for _, tt := range valuetests {
		actual, err := tt.input.Value()
		if err != nil {
			t.Errorf(
				`Option.Value() for %s: unexpected err: %s`,
				tt.desc, err,
			)
			continue
		}

		if tt.expected != actual {
			t.Errorf(
				`Option.Value() for %s: expected "%s", actual "%s"`,
				tt.desc, tt.expected, actual,
			)
		}
	}
}
func TestOption_Value_default_and_command(t *testing.T) {
	option := Option{content: content{Default: "foo", Command: "echo bar"}}
	_, err := option.Value()
	if err == nil {
		t.Fatalf(
			"option.Value() for %s: expected err, actual nil",
			"both Default and Command defined",
		)
	}
}

func TestOption_Value_private_and_environment(t *testing.T) {
	option := Option{Private: true, Environment: "OPTION_VAR"}
	_, err := option.Value()
	if err == nil {
		t.Fatalf(
			"option.Value() for %s: expected err, actual nil",
			"both Private and Environment variable defined",
		)
	}
}
