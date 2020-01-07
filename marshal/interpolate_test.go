package marshal

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestInterpolate_string(t *testing.T) {
	values := map[string]string{"name": "foo", "other": "bar"}

	input := "My name is ${name}, not ${invalid}"
	want := "My name is foo, not ${invalid}"

	err := Interpolate(&input, values)
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal(want, input))
}

func TestInterpolate_slice(t *testing.T) {
	values := map[string]string{"name": "foo", "other": "bar"}

	input := []string{"My name", "is ${name}", "not ${invalid}"}
	want := []string{"My name", "is foo", "not ${invalid}"}

	err := Interpolate(&input, values)
	assert.NilError(t, err)

	assert.Check(t, cmp.DeepEqual(want, input))
}

func TestInterpolate_struct(t *testing.T) {
	values := map[string]string{"name": "foo", "other": "bar"}

	type s struct {
		Name string
		Not  string
	}

	input := s{"it's ${name}", "not ${invalid}"}
	want := s{"it's foo", "not ${invalid}"}

	err := Interpolate(&input, values)
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal(want, input))
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
			escaped := escape([]byte(tt.input))
			got := string(escaped)

			if tt.want != got {
				t.Errorf("want %q, got %q", tt.want, got)
			}
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
			actual, err := mapInterpolate([]byte(tt.input), vars)
			assert.NilError(t, err)

			assert.Check(t, cmp.Equal(tt.want, string(actual)))
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
			got := FindPotentialVariables([]byte(tt.input))
			assert.Check(t, cmp.DeepEqual(tt.want, got))
		})
	}
}
