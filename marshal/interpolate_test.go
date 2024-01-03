package marshal

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
)

func TestInterpolate_string(t *testing.T) {
	g := ghost.New(t)

	values := map[string]string{"name": "foo", "other": "bar"}

	input := "My name is ${name}, not ${invalid}"
	want := "My name is foo, not ${invalid}"

	err := Interpolate(&input, values)
	g.NoError(err)

	g.Should(be.Equal(input, want))
}

func TestInterpolate_slice(t *testing.T) {
	g := ghost.New(t)

	values := map[string]string{"name": "foo", "other": "bar"}

	input := []string{"My name", "is ${name}", "not ${invalid}"}
	want := []string{"My name", "is foo", "not ${invalid}"}

	err := Interpolate(&input, values)
	g.NoError(err)

	g.Should(be.DeepEqual(input, want))
}

func TestInterpolate_struct(t *testing.T) {
	g := ghost.New(t)

	values := map[string]string{"name": "foo", "other": "bar"}

	type s struct {
		Name string
		Not  string
	}

	input := s{"it's ${name}", "not ${invalid}"}
	want := s{"it's foo", "not ${invalid}"}

	err := Interpolate(&input, values)
	g.NoError(err)

	g.Should(be.Equal(input, want))
}

func TestEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"$", "$"},
		{"$$", "$"},
		{"$$$", "$$"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			g := ghost.New(t)

			escaped := escape([]byte(tt.input))
			got := string(escaped)

			g.Should(be.Equal(got, tt.want))
		})
	}
}

func TestMap(t *testing.T) {
	vars := map[string]string{"foo": "bar"}

	tests := []struct {
		input string
		want  string
	}{
		{"${foo}", "bar"},
		{"foo", "foo"},
		{"$foo", "$foo"},
		{"${foo}${foo}", "barbar"},
		{"${foo}${bar}", "bar${bar}"},
		{"$${foo}", "$${foo}"},
		{"$$${foo}", "$$bar"},
		{"$", "$"},
		{"$$", "$$"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			g := ghost.New(t)

			got, err := mapInterpolate([]byte(tt.input), vars)
			g.NoError(err)

			g.Should(be.Equal(string(got), tt.want))
		})
	}
}

func TestFindPotentialVariables(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", []string{}},
		{"${}", []string{}},
		{"foo", []string{}},
		{"$foo", []string{}},
		{"${foo}", []string{"foo"}},
		{"${f-o-o}", []string{"f-o-o"}},
		{"${f_o_o}", []string{"f_o_o"}},
		{"${foo}${bar}", []string{"foo", "bar"}},
		{"${foo}${FOO}", []string{"foo", "FOO"}},
		{"_-${foo}.  ${bar} baz", []string{"foo", "bar"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			g := ghost.New(t)

			got := FindPotentialVariables([]byte(tt.input))
			g.Should(be.DeepEqual(got, tt.want))
		})
	}
}
