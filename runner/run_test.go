package runner

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
			var got Run

			if err := yaml.UnmarshalStrict([]byte(tt.yaml), &got); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatched values:\n%s", diff)
			}
		})
	}
}

var environmentTests = []struct {
	input          string
	expectedLength int
}{
	{`{}`, 0},
	{`{set-environment: {}}`, 0},
	{`{set-environment: {foo: bar}}`, 1},
	{`{set-environment: {foo: bar, bar: baz}}`, 2},
}

func TestRun_UnmarshalYAML_SetEnvironment(t *testing.T) {
	for _, testCase := range environmentTests {
		r := Run{}

		if err := yaml.UnmarshalStrict([]byte(testCase.input), &r); err != nil {
			t.Errorf(
				"yaml.UnmarshalStrict(%s, ...): unexpected error: %s",
				testCase.input, err,
			)
			continue
		}

		if testCase.expectedLength != len(r.SetEnvironment) {
			t.Errorf(
				"yaml.UnmarshalStrict(%s, ...): expected %d environment items, got %d",
				testCase.input, testCase.expectedLength, len(r.SetEnvironment),
			)
		}
	}
}

var multipleActionTests = []string{
	`{command: example, task: echo 'hello'}`,
	`{command: example, environment: {foo: bar}}`,
	`{task: echo 'hello', environment: {foo: bar}}`,
	`{command: example, task: echo 'hello', environment: {foo: bar}}`,
	`{environment: {foo: bar}, set-environment: {bar: baz}}`,
}

func TestRun_UnmarshalYAML_command_and_subtask(t *testing.T) {
	for _, input := range multipleActionTests {
		r := Run{}
		if err := yaml.UnmarshalStrict([]byte(input), &r); err == nil {
			t.Errorf(
				"yaml.UnmarshalStrict(%s, ...): expected error, received nil",
				input,
			)
		}
	}
}

var shouldtests = []struct {
	desc     string
	input    Run
	expected bool
	vars     map[string]string
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

func TestRun_shouldRun(t *testing.T) {
	for _, tt := range shouldtests {
		actual, err := tt.input.shouldRun(tt.vars)
		if err != nil {
			t.Errorf(
				"task.shouldRun() for %s: unexpected error: %s",
				tt.desc, err,
			)
			continue
		}
		if tt.expected != actual {
			t.Errorf(
				"task.shouldRun() for %s: expected: %t, actual: %t",
				tt.desc, tt.expected, actual,
			)
		}
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
			var got RunList

			if err := yaml.UnmarshalStrict([]byte(tt.yaml), &got); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatched values:\n%s", diff)
			}
		})
	}
}
