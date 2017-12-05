package option

import (
	"os"
	"reflect"
	"testing"

	"github.com/rliebz/tusk/config/marshal"
	"github.com/rliebz/tusk/config/when"
	"github.com/rliebz/tusk/config/whentest"
	yaml "gopkg.in/yaml.v2"
)

func TestOption_Dependencies(t *testing.T) {
	option := &Option{DefaultValues: valueList{
		{When: whentest.False, Value: "foo"},
		{When: when.When{
			Equal: map[string]marshal.StringList{
				"foo": {"foovalue"},
				"bar": {"barvalue"},
			},
		}, Value: "bar"},
		{When: when.When{
			NotEqual: map[string]marshal.StringList{
				"baz": {"bazvalue"},
			},
		}, Value: "bar"},
	}}

	expected := []string{"foo", "bar", "baz"}
	actual := option.Dependencies()
	if !equalUnordered(expected, actual) {
		t.Errorf(
			"Option.Dependencies(): expected %s, actual %s",
			expected, actual,
		)
	}
}

// nolint: dupl
func equalUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]interface{})
	for _, val := range a {
		aMap[val] = struct{}{}
	}

	bMap := make(map[string]interface{})
	for _, val := range b {
		bMap[val] = struct{}{}
	}

	return reflect.DeepEqual(aMap, bMap)
}

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
		&Option{DefaultValues: valueList{
			{Value: "default"},
		}},
		"default",
	},
	{
		"command only",
		&Option{DefaultValues: valueList{
			{Command: "echo command"},
		}},
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
		"conditional value",
		&Option{DefaultValues: valueList{
			{When: whentest.False, Value: "foo"},
			{When: whentest.True, Value: "bar"},
			{When: whentest.False, Value: "baz"},
		}},
		"bar",
	},
	{
		"passed when all settings are defined",
		&Option{
			Environment: "OPTION_VAR",
			DefaultValues: valueList{
				{When: whentest.True, Value: "when"},
			},
			Passed: "passed",
		},
		"passed",
	},
}

func TestOption_Evaluate(t *testing.T) {
	if err := os.Setenv("OPTION_VAR", "option_val"); err != nil {
		t.Fatalf("unexpected err setting environment variable: %s", err)
	}

	for _, tt := range valuetests {
		actual, err := tt.input.Evaluate()
		if err != nil {
			t.Errorf(
				`Option.Evaluate() for %s: unexpected err: %s`,
				tt.desc, err,
			)
			continue
		}

		if tt.expected != actual {
			t.Errorf(
				`Option.Evaluate() for %s: expected "%s", actual "%s"`,
				tt.desc, tt.expected, actual,
			)
		}
	}
}

func TestOption_Evaluate_sets_environment_variable(t *testing.T) {
	expected := "test value"
	envName := "EVALUATE_OUTPUT_VAR"
	o := Option{
		Passed: expected,
		Export: envName,
	}

	if err := os.Unsetenv(envName); err != nil {
		t.Errorf(`os.Unsetenv(%s): unexpected err: %s`, envName, err)
	}

	if _, err := o.Evaluate(); err != nil {
		t.Errorf(`Option.Evaluate(): unexpected err: %s`, err)
	}

	if actual := os.Getenv(envName); actual != expected {
		t.Errorf(
			`Option.Evaluate() exported var "%s": expected "%s", actual "%s"`,
			envName, expected, actual,
		)
	}
}

func TestOption_Evaluate_required_nothing_passed(t *testing.T) {
	option := Option{Required: true}

	if _, err := option.Evaluate(); err == nil {
		t.Fatal(
			"Option.Evaluate() for required option: expected err, actual nil",
		)
	}
}

func TestOption_Evaluate_required_with_passed(t *testing.T) {
	expected := "foo"
	option := Option{Required: true, Passed: expected}

	actual, err := option.Evaluate()
	if err != nil {
		t.Fatalf("Option.Evaluate(): unexpected error: %s", err)
	}

	if expected != actual {
		t.Errorf(
			`Option.Evaluate(): expected "%s", actual "%s"`,
			expected, actual,
		)
	}
}

func TestOption_Evaluate_required_with_environment(t *testing.T) {
	envVar := "OPTION_VAR"
	expected := "foo"

	option := Option{Required: true, Environment: envVar}
	if err := os.Setenv(envVar, expected); err != nil {
		t.Fatalf("unexpected err setting environment variable: %s", err)
	}

	actual, err := option.Evaluate()
	if err != nil {
		t.Fatalf("Option.Evaluate(): unexpected error: %s", err)
	}

	if expected != actual {
		t.Errorf(
			`Option.Evaluate(): expected "%s", actual "%s"`,
			expected, actual,
		)
	}
}

var evaluteTypeDefaultTests = []struct {
	typeName string
	expected string
}{
	{"int", "0"},
	{"INTEGER", "0"},
	{"Float", "0"},
	{"float64", "0"},
	{"double", "0"},
	{"bool", "false"},
	{"boolean", "false"},
	{"", ""},
}

func TestOption_Evaluate_type_defaults(t *testing.T) {
	for _, tt := range evaluteTypeDefaultTests {
		opt := Option{Type: tt.typeName}
		actual, err := opt.Evaluate()
		if err != nil {
			t.Errorf("Option.Evaluate(): unexpected error: %s", err)
			continue
		}

		if tt.expected != actual {
			t.Errorf(
				`Option.Evaluate(): expected "%s", actual "%s"`,
				tt.expected, actual,
			)
		}
	}
}

func TestOption_UnmarshalYAML(t *testing.T) {
	s := []byte(`{usage: foo, name: ignored}`)
	expected := Option{
		Usage: "foo",
		Name:  "",
	}
	actual := Option{}

	if err := yaml.Unmarshal(s, &actual); err != nil {
		t.Fatalf("yaml.Unmarshal(%s, ...): unexpected error: %s", s, err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf(
			`yaml.Unmarshal(%s, ...): expected "%#v", actual "%#v"`,
			s, expected, actual,
		)
	}

}

var unmarshalOptionErrorTests = []struct {
	desc  string
	input string
}{
	{
		"invalid option definition",
		"string only",
	},
	{
		"short name exceeds one character",
		"{short: foo}",
	},
	{
		"private and required defined",
		"{private: true, required: true}",
	},
	{
		"private and environment defined",
		"{private: true, environment: ENV_VAR}",
	},
	{
		"required and default defined",
		"{required: true, default: foo}",
	},
}

func TestOption_UnmarshalYAML_invalid_definitions(t *testing.T) {
	for _, tt := range unmarshalOptionErrorTests {
		o := Option{}
		if err := yaml.Unmarshal([]byte(tt.input), &o); err == nil {
			t.Errorf(
				"yaml.Unmarshal(%s, ...): expected error for %s, actual nil",
				tt.input, tt.desc,
			)
		}
	}
}
