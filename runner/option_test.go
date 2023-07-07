package runner

import (
	"reflect"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	"github.com/rliebz/tusk/marshal"
	yaml "gopkg.in/yaml.v2"
)

func TestOption_Dependencies(t *testing.T) {
	g := ghost.New(t)

	option := &Option{DefaultValues: ValueList{
		{When: WhenList{whenFalse}, Value: "foo"},
		{When: WhenList{createWhen(
			withWhenEqual("foo", "foovalue"),
			withWhenEqual("bar", "barvalue"),
		)}, Value: "bar"},
		{When: WhenList{createWhen(
			withWhenNotEqual("baz", "bazvalue"),
		)}, Value: "bar"},
	}}

	g.Should(be.True(equalUnordered([]string{"foo", "bar", "baz"}, option.Dependencies())))
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

func TestOption_Evaluate(t *testing.T) {
	t.Setenv("OPTION_VAR", "option_val")

	// Env var `OPTION_VAR` will be set to `option_val`
	tests := []struct {
		name  string
		input *Option
		want  string
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
			&Option{Passable: Passable{Passed: "passed"}},
			"passed",
		},
		{
			"conditional value",
			&Option{DefaultValues: ValueList{
				{When: WhenList{whenFalse}, Value: "foo"},
				{When: WhenList{whenTrue}, Value: "bar"},
				{When: WhenList{whenFalse}, Value: "baz"},
			}},
			"bar",
		},
		{
			"passed when all settings are defined",
			&Option{
				Environment: "OPTION_VAR",
				DefaultValues: ValueList{
					{When: WhenList{whenTrue}, Value: "when"},
				},
				Passable: Passable{
					Passed: "passed",
				},
			},
			"passed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			got, err := tt.input.Evaluate(Context{}, nil)
			g.NoError(err)

			g.Should(be.Equal(tt.want, got))
		})
	}
}

func TestOption_Evaluate_required_nothing_passed(t *testing.T) {
	g := ghost.New(t)

	option := Option{
		Passable: Passable{
			Name: "my-opt",
		},
		Required: true,
	}

	_, err := option.Evaluate(Context{}, nil)
	g.Should(be.ErrorEqual("no value passed for required option: my-opt", err))
}

func TestOption_Evaluate_passes_vars(t *testing.T) {
	g := ghost.New(t)

	want := "some value"
	opt := Option{
		DefaultValues: ValueList{
			{When: WhenList{whenFalse}, Value: "wrong"},
			{
				When:  WhenList{createWhen(withWhenEqual("foo", "foovalue"))},
				Value: want,
			},
			{When: WhenList{whenFalse}, Value: "oops"},
		},
	}

	got, err := opt.Evaluate(Context{}, map[string]string{"foo": "foovalue"})
	g.NoError(err)

	g.Should(be.Equal(want, got))
}

func TestOption_Evaluate_required_with_passed(t *testing.T) {
	g := ghost.New(t)

	want := "foo"
	option := Option{
		Required: true,
		Passable: Passable{
			Passed: want,
		},
	}

	got, err := option.Evaluate(Context{}, nil)
	g.NoError(err)

	g.Should(be.Equal(want, got))
}

func TestOption_Evaluate_required_with_environment(t *testing.T) {
	g := ghost.New(t)

	envVar := "OPTION_VAR"
	want := "foo"

	option := Option{Required: true, Environment: envVar}
	t.Setenv(envVar, want)

	got, err := option.Evaluate(Context{}, nil)
	g.NoError(err)

	g.Should(be.Equal(want, got))
}

func TestOption_Evaluate_values_none_specified(t *testing.T) {
	g := ghost.New(t)

	want := ""
	option := Option{
		Passable: Passable{
			ValuesAllowed: marshal.StringList{"red", "herring"},
		},
	}

	got, err := option.Evaluate(Context{}, nil)
	g.NoError(err)

	g.Should(be.Equal(want, got))
}

func TestOption_Evaluate_values_with_passed(t *testing.T) {
	g := ghost.New(t)

	want := "foo"
	option := Option{
		Passable: Passable{
			Passed:        want,
			ValuesAllowed: marshal.StringList{"red", want, "herring"},
		},
	}

	got, err := option.Evaluate(Context{}, nil)
	g.NoError(err)

	g.Should(be.Equal(want, got))
}

func TestOption_Evaluate_values_with_environment(t *testing.T) {
	g := ghost.New(t)

	envVar := "OPTION_VAR"
	want := "foo"

	option := Option{
		Environment: envVar,
		Passable: Passable{
			ValuesAllowed: marshal.StringList{"red", want, "herring"},
		},
	}

	t.Setenv(envVar, want)

	got, err := option.Evaluate(Context{}, nil)
	g.NoError(err)

	g.Should(be.Equal(want, got))
}

func TestOption_Evaluate_values_with_invalid_passed(t *testing.T) {
	g := ghost.New(t)

	want := "foo"
	option := Option{
		Passable: Passable{
			Name:          "my-opt",
			Passed:        want,
			ValuesAllowed: marshal.StringList{"bad", "values", "FOO"},
		},
	}

	_, err := option.Evaluate(Context{}, nil)
	g.Should(be.ErrorEqual(`value "foo" for option "my-opt" must be one of [bad values FOO]`, err))
}

func TestOption_Evaluate_values_with_invalid_environment(t *testing.T) {
	g := ghost.New(t)

	envVar := "OPTION_VAR"
	want := "foo"

	option := Option{
		Environment: envVar,
		Passable: Passable{
			Name:          "my-opt",
			ValuesAllowed: marshal.StringList{"bad", "values", "FOO"},
		},
	}

	t.Setenv(envVar, want)

	_, err := option.Evaluate(Context{}, nil)
	g.Should(be.ErrorEqual(`value "foo" for option "my-opt" must be one of [bad values FOO]`, err))
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
		opt := Option{
			Passable: Passable{
				Type: tt.typeName,
			},
		}
		actual, err := opt.Evaluate(Context{}, nil)
		if err != nil {
			t.Errorf("Option.Evaluate(): unexpected error: %s", err)
			continue
		}

		if tt.expected != actual {
			t.Errorf(
				"Option.Evaluate(): expected %q, actual %q",
				tt.expected, actual,
			)
		}
	}
}

func TestOption_UnmarshalYAML(t *testing.T) {
	g := ghost.New(t)

	s := []byte(`{usage: foo, values: [foo, bar]}`)
	want := Option{
		Passable: Passable{
			Name:          "",
			Usage:         "foo",
			ValuesAllowed: []string{"foo", "bar"},
		},
	}

	var got Option
	err := yaml.UnmarshalStrict(s, &got)
	g.NoError(err)

	g.Should(be.DeepEqual(want, got))
}

func TestOption_UnmarshalYAML_invalid_definitions(t *testing.T) {
	tests := []struct {
		name  string
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			var o Option
			err := yaml.UnmarshalStrict([]byte(tt.input), &o)
			g.Should(be.Error(err))
		})
	}
}

func TestGetOptionsWithOrder(t *testing.T) {
	g := ghost.New(t)

	ms := yaml.MapSlice{
		{Key: "foo", Value: &Option{Environment: "fooenv"}},
		{Key: "bar", Value: &Option{Environment: "barenv"}},
	}

	options, err := getOptionsWithOrder(ms)
	g.NoError(err)

	g.Must(be.SliceLen(2, options))

	g.Should(be.Equal("foo", options[0].Name))
	g.Should(be.Equal("fooenv", options[0].Environment))
	g.Should(be.Equal("bar", options[1].Name))
	g.Should(be.Equal("barenv", options[1].Environment))
}
