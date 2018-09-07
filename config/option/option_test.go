package option

import (
	"os"
	"reflect"
	"testing"

	"github.com/rliebz/tusk/config/marshal"
	"github.com/rliebz/tusk/config/when"
	yaml "gopkg.in/yaml.v2"
)

func TestOption_Dependencies(t *testing.T) {
	option := &Option{DefaultValues: ValueList{
		{When: when.List{when.False}, Value: "foo"},
		{When: when.List{when.Create(
			when.WithEqual("foo", "foovalue"),
			when.WithEqual("bar", "barvalue"),
		)}, Value: "bar"},
		{When: when.List{when.Create(
			when.WithNotEqual("baz", "bazvalue"),
		)}, Value: "bar"},
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
		&Option{DefaultValues: ValueList{
			{Value: "default"},
		}},
		"default",
	},
	{
		"command only",
		&Option{DefaultValues: ValueList{
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
		&Option{DefaultValues: ValueList{
			{When: when.List{when.False}, Value: "foo"},
			{When: when.List{when.True}, Value: "bar"},
			{When: when.List{when.False}, Value: "baz"},
		}},
		"bar",
	},
	{
		"passed when all settings are defined",
		&Option{
			Environment: "OPTION_VAR",
			DefaultValues: ValueList{
				{When: when.List{when.True}, Value: "when"},
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
		actual, err := tt.input.Evaluate(nil)
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

func TestOption_Evaluate_required_nothing_passed(t *testing.T) {
	option := Option{Required: true}

	if _, err := option.Evaluate(nil); err == nil {
		t.Fatal(
			"Option.Evaluate() for required option: expected err, actual nil",
		)
	}
}

func TestOption_Evaluate_passes_vars(t *testing.T) {
	expected := "some value"
	opt := Option{
		DefaultValues: ValueList{
			{When: when.List{when.False}, Value: "wrong"},
			{
				When:  when.List{when.Create(when.WithEqual("foo", "foovalue"))},
				Value: expected,
			},
			{When: when.List{when.False}, Value: "oops"},
		},
	}

	actual, err := opt.Evaluate(map[string]string{"foo": "foovalue"})
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

func TestOption_Evaluate_required_with_passed(t *testing.T) {
	expected := "foo"
	option := Option{Required: true, Passed: expected}

	actual, err := option.Evaluate(nil)
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

	actual, err := option.Evaluate(nil)
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

func TestOption_Evaluate_values_none_specified(t *testing.T) {
	expected := ""
	option := Option{
		valueWithList: valueWithList{
			ValuesAllowed: marshal.StringList{"red", "herring"},
		},
	}

	actual, err := option.Evaluate(nil)
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

func TestOption_Evaluate_values_with_passed(t *testing.T) {
	expected := "foo"
	option := Option{
		Passed: expected,
		valueWithList: valueWithList{
			ValuesAllowed: marshal.StringList{"red", expected, "herring"},
		},
	}

	actual, err := option.Evaluate(nil)
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

func TestOption_Evaluate_values_with_environment(t *testing.T) {
	envVar := "OPTION_VAR"
	expected := "foo"

	option := Option{
		Environment: envVar,
		valueWithList: valueWithList{
			ValuesAllowed: marshal.StringList{"red", expected, "herring"},
		},
	}

	if err := os.Setenv(envVar, expected); err != nil {
		t.Fatalf("unexpected err setting environment variable: %s", err)
	}

	actual, err := option.Evaluate(nil)
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

func TestOption_Evaluate_values_with_invalid_passed(t *testing.T) {
	expected := "foo"
	option := Option{
		Passed: expected,
		valueWithList: valueWithList{
			ValuesAllowed: marshal.StringList{"bad", "values", "FOO"},
		},
	}

	_, err := option.Evaluate(nil)
	if err == nil {
		t.Fatalf(
			"Option.Evaluate(): expected error for invalid passed value, got nil",
		)
	}
}

func TestOption_Evaluate_values_with_invalid_environment(t *testing.T) {
	envVar := "OPTION_VAR"
	expected := "foo"

	option := Option{
		Environment: envVar,
		valueWithList: valueWithList{
			ValuesAllowed: marshal.StringList{"bad", "values", "FOO"},
		},
	}

	if err := os.Setenv(envVar, expected); err != nil {
		t.Fatalf("unexpected err setting environment variable: %s", err)
	}

	_, err := option.Evaluate(nil)
	if err == nil {
		t.Fatalf(
			"Option.Evaluate(): expected error for invalid environment value, got nil",
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
		actual, err := opt.Evaluate(nil)
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
	s := []byte(`{usage: foo, values: [foo, bar], name: ignored}`)
	expected := Option{
		Usage: "foo",
		valueWithList: valueWithList{
			ValuesAllowed: []string{"foo", "bar"},
		},
		Name: "",
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
		"private and values defined",
		"{private: true, values: [foo, bar]}",
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

// nolint: dupl
func TestGetOptionsWithOrder(t *testing.T) {
	name := "foo"
	env := "fooenv"
	ms := yaml.MapSlice{
		{Key: name, Value: &Option{Environment: env}},
		{Key: "bar", Value: &Option{Environment: "barenv"}},
	}

	options, ordered, err := GetOptionsWithOrder(ms)
	if err != nil {
		t.Fatalf("GetOptionsWithOrder(ms) => unexpected error: %v", err)
	}

	if len(ms) != len(options) || len(ms) != len(ordered) {
		t.Fatalf(
			"GetOptionsWithOrder(ms) => want %d items, got %d in map and %d in slice",
			len(ms), len(options), len(ordered),
		)
	}

	opt, ok := options[name]
	if !ok {
		t.Fatalf("GetOptionsWithOrder(ms) => item %q is not in map", name)
	}

	if name != opt.Name {
		t.Errorf(
			"GetOptionsWithOrder(ms) => want opt.Name %q, got %q",
			name, opt.Name,
		)
	}

	if env != opt.Environment {
		t.Errorf(
			"GetOptionsWithOrder(ms) => want opt.Environment %q, got %q",
			env, opt.Environment,
		)
	}

	expectedOrder := []string{"foo", "bar"}
	if !reflect.DeepEqual(expectedOrder, ordered) {
		t.Errorf(
			"GetOptionsWithOrder(ms) => want ordered %v, got %v",
			expectedOrder, ordered,
		)
	}

}
