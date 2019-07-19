package appcli

import (
	"fmt"
	"testing"

	"github.com/rliebz/tusk/config"
)

var flagPrefixerTests = []struct {
	flags       string
	placeholder string
	expected    string
}{
	{"a", "", "-a"},
	{"a", "foo", "-a <foo>"},
	{"aa", "", "    --aa"},
	{"aa", "foo", "    --aa <foo>"},
	{"a, aa", "", "-a, --aa"},
	{"aa, a", "", "-a, --aa"},
	{"a, aa", "foo", "-a, --aa <foo>"},
}

func TestFlagPrefixer(t *testing.T) {
	for _, tt := range flagPrefixerTests {
		actual := flagPrefixer(tt.flags, tt.placeholder)
		if tt.expected != actual {
			t.Errorf(
				"flagPrefixer(%q, %q): expected %q, got %q",
				tt.flags, tt.placeholder, tt.expected, actual,
			)
		}
	}
}

var argsSectionTests = []struct {
	desc     string
	taskCfg  string
	expected string
}{
	{
		"no args",
		"",
		"",
	},
	{
		"one args",
		"foo: {usage: 'some usage'}",
		`

Arguments:
   foo  some usage`,
	},
	{
		"args without usage",
		"foo: {}, bar: {}",
		`

Arguments:
   foo
   bar`,
	},
	{
		"args with usage",
		"foo: {usage: 'some usage'}, bar: {usage: 'other usage'}",
		`

Arguments:
   foo  some usage
   bar  other usage`,
	},
	{
		"variable length arguments",
		"a: {usage: 'some usage'}, aaaaa: {usage: 'other usage'}",
		`

Arguments:
   a      some usage
   aaaaa  other usage`,
	},
}

func TestCreateArgsSection(t *testing.T) {
	for _, tt := range argsSectionTests {
		taskName := "someTaskName"
		t.Run(tt.desc, func(t *testing.T) {
			cfgText := fmt.Sprintf("tasks: { %s: { args: {%s} } }", taskName, tt.taskCfg)
			cfg, err := config.Parse([]byte(cfgText))
			if err != nil {
				t.Fatal(err)
			}

			actual := createArgsSection(cfg.Tasks[taskName])
			if tt.expected != actual {
				t.Errorf(
					"want %q, got %q", tt.expected, actual,
				)
			}
		})

	}
}
