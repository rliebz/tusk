package runner

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	yaml "gopkg.in/yaml.v2"

	"github.com/rliebz/tusk/marshal"
)

func TestOption_Dependencies(t *testing.T) {
	g := ghost.New(t)

	option := &Option{DefaultValues: marshal.Slice[Value]{
		{When: WhenList{whenFalse}, Value: "foo"},
		{When: WhenList{createWhen(
			withWhenEqual("foo", "foovalue"),
			withWhenEqual("bar", "barvalue"),
		)}, Value: "bar"},
		{When: WhenList{createWhen(
			withWhenNotEqual("baz", "bazvalue"),
		)}, Value: "bar"},
	}}

	g.Should(beEqualUnordered([]string{"foo", "bar", "baz"}, option.Dependencies()))
}

func beEqualUnordered[T comparable](a, b []T) ghost.Result {
	aMap := make(map[T]interface{})
	for _, val := range a {
		aMap[val] = struct{}{}
	}

	bMap := make(map[T]interface{})
	for _, val := range b {
		bMap[val] = struct{}{}
	}

	return be.DeepEqual(aMap, bMap)
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
			&Option{DefaultValues: marshal.Slice[Value]{
				{Value: "default"},
			}},
			"default",
		},
		{
			"command only",
			&Option{DefaultValues: marshal.Slice[Value]{
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
			&Option{DefaultValues: marshal.Slice[Value]{
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
				DefaultValues: marshal.Slice[Value]{
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

			g.Should(be.Equal(got, tt.want))
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
	g.Should(be.ErrorEqual(err, "no value passed for required option: my-opt"))
}

func TestOption_Evaluate_passes_vars(t *testing.T) {
	g := ghost.New(t)

	want := "some value"
	opt := Option{
		DefaultValues: marshal.Slice[Value]{
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

	g.Should(be.Equal(got, want))
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

	g.Should(be.Equal(got, want))
}

func TestOption_Evaluate_required_with_environment(t *testing.T) {
	g := ghost.New(t)

	envVar := "OPTION_VAR"
	want := "foo"

	option := Option{Required: true, Environment: envVar}
	t.Setenv(envVar, want)

	got, err := option.Evaluate(Context{}, nil)
	g.NoError(err)

	g.Should(be.Equal(got, want))
}

func TestOption_Evaluate_values_none_specified(t *testing.T) {
	g := ghost.New(t)

	want := ""
	option := Option{
		Passable: Passable{
			ValuesAllowed: marshal.Slice[string]{"red", "herring"},
		},
	}

	got, err := option.Evaluate(Context{}, nil)
	g.NoError(err)

	g.Should(be.Equal(got, want))
}

func TestOption_Evaluate_values_with_passed(t *testing.T) {
	g := ghost.New(t)

	want := "foo"
	option := Option{
		Passable: Passable{
			Passed:        want,
			ValuesAllowed: marshal.Slice[string]{"red", want, "herring"},
		},
	}

	got, err := option.Evaluate(Context{}, nil)
	g.NoError(err)

	g.Should(be.Equal(got, want))
}

func TestOption_Evaluate_values_with_environment(t *testing.T) {
	g := ghost.New(t)

	envVar := "OPTION_VAR"
	want := "foo"

	option := Option{
		Environment: envVar,
		Passable: Passable{
			ValuesAllowed: marshal.Slice[string]{"red", want, "herring"},
		},
	}

	t.Setenv(envVar, want)

	got, err := option.Evaluate(Context{}, nil)
	g.NoError(err)

	g.Should(be.Equal(got, want))
}

func TestOption_Evaluate_values_with_invalid_passed(t *testing.T) {
	g := ghost.New(t)

	want := "foo"
	option := Option{
		Passable: Passable{
			Name:          "my-opt",
			Passed:        want,
			ValuesAllowed: marshal.Slice[string]{"bad", "values", "FOO"},
		},
	}

	_, err := option.Evaluate(Context{}, nil)
	g.Should(be.ErrorEqual(err, `value "foo" for option "my-opt" must be one of [bad values FOO]`))
}

func TestOption_Evaluate_values_with_invalid_environment(t *testing.T) {
	g := ghost.New(t)

	envVar := "OPTION_VAR"
	want := "foo"

	option := Option{
		Environment: envVar,
		Passable: Passable{
			Name:          "my-opt",
			ValuesAllowed: marshal.Slice[string]{"bad", "values", "FOO"},
		},
	}

	t.Setenv(envVar, want)

	_, err := option.Evaluate(Context{}, nil)
	g.Should(be.ErrorEqual(err, `value "foo" for option "my-opt" must be one of [bad values FOO]`))
}

func TestOption_Evaluate_type_defaults(t *testing.T) {
	tests := []struct {
		typeName string
		want     string
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

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			g := ghost.New(t)

			opt := Option{
				Passable: Passable{
					Type: tt.typeName,
				},
			}

			got, err := opt.Evaluate(Context{}, nil)
			g.NoError(err)

			g.Should(be.Equal(got, tt.want))
		})
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

	g.Should(be.DeepEqual(got, want))
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
			"private and short defined",
			"{private: true, short: n}",
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

	g.Must(be.SliceLen(options, 2))

	g.Should(be.Equal(options[0].Name, "foo"))
	g.Should(be.Equal(options[0].Environment, "fooenv"))
	g.Should(be.Equal(options[1].Name, "bar"))
	g.Should(be.Equal(options[1].Environment, "barenv"))
}
