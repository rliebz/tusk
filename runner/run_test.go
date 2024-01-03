package runner

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	yaml "gopkg.in/yaml.v2"
)

func TestRun_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want Run
	}{
		{
			"short-command",
			`example`,
			Run{
				Command: CommandList{{Exec: "example", Print: "example"}},
			},
		},
		{
			"short-command-list",
			`[one,two]`,
			Run{
				Command: CommandList{
					{Exec: "one", Print: "one"},
					{Exec: "two", Print: "two"},
				},
			},
		},
		{
			"named-command",
			`command: example`,
			Run{
				Command: CommandList{{Exec: "example", Print: "example"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			var got Run
			err := yaml.UnmarshalStrict([]byte(tt.yaml), &got)
			g.NoError(err)

			g.Should(be.DeepEqual(got, tt.want))
		})
	}
}

func TestRun_UnmarshalYAML_SetEnvironment(t *testing.T) {
	tests := []struct {
		input   string
		wantLen int
	}{
		{`{}`, 0},
		{`{set-environment: {}}`, 0},
		{`{set-environment: {foo: bar}}`, 1},
		{`{set-environment: {foo: bar, bar: baz}}`, 2},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			g := ghost.New(t)

			var r Run
			err := yaml.UnmarshalStrict([]byte(tt.input), &r)
			g.NoError(err)

			g.Should(be.MapLen(r.SetEnvironment, tt.wantLen))
		})
	}
}

func TestRun_UnmarshalYAML_command_and_subtask(t *testing.T) {
	tests := []string{
		`{command: example, task: echo 'hello'}`,
		`{command: example, environment: {foo: bar}}`,
		`{task: echo 'hello', environment: {foo: bar}}`,
		`{command: example, task: echo 'hello', environment: {foo: bar}}`,
		`{environment: {foo: bar}, set-environment: {bar: baz}}`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			g := ghost.New(t)

			var r Run
			err := yaml.UnmarshalStrict([]byte(input), &r)
			g.Should(be.Error(err))
		})
	}
}

func TestRun_shouldRun(t *testing.T) {
	tests := []struct {
		name  string
		input Run
		want  bool
		vars  map[string]string
	}{
		{"no when clause", Run{}, true, nil},
		{"true when clause", Run{When: WhenList{whenTrue}}, true, nil},
		{"false when clause", Run{When: WhenList{whenFalse}}, false, nil},
		{
			"var matches condition",
			Run{When: WhenList{createWhen(withWhenEqual("foo", "bar"))}},
			true,
			map[string]string{"foo": "bar"},
		},
		{
			"var does not match condition",
			Run{When: WhenList{createWhen(withWhenEqual("foo", "bar"))}},
			false,
			map[string]string{"foo": "baz"},
		},
		{
			"var was not passed",
			Run{When: WhenList{createWhen(withWhenEqual("foo", "bar"))}},
			false,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			got, err := tt.input.shouldRun(Context{}, tt.vars)
			g.NoError(err)

			g.Should(be.Equal(got, tt.want))
		})
	}
}

func TestRunList_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want RunList
	}{
		{
			"single-short-run",
			`example`,
			RunList{
				{Command: CommandList{{Exec: "example", Print: "example"}}},
			},
		},
		{
			"list-short-runs",
			`[one,two]`,
			RunList{
				{Command: CommandList{{Exec: "one", Print: "one"}}},
				{Command: CommandList{{Exec: "two", Print: "two"}}},
			},
		},
		{
			"list-full-runs",
			`[{command: foo},{set-environment: {bar: null}}]`,
			RunList{
				{Command: CommandList{{Exec: "foo", Print: "foo"}}},
				{SetEnvironment: map[string]*string{"bar": nil}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			var got RunList
			err := yaml.UnmarshalStrict([]byte(tt.yaml), &got)
			g.NoError(err)

			g.Should(be.DeepEqual(got, tt.want))
		})
	}
}
