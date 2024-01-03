package appcli

import (
	"fmt"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"

	"github.com/rliebz/tusk/runner"
)

func TestFlagPrefixer(t *testing.T) {
	tests := []struct {
		name        string
		flags       string
		placeholder string
		want        string
	}{
		{"short", "a", "", "-a"},
		{"short placeholder", "a", "foo", "-a <foo>"},
		{"long", "aa", "", "    --aa"},
		{"long placeholder", "aa", "foo", "    --aa <foo>"},
		{"short first", "a, aa", "", "-a, --aa"},
		{"long first", "aa, a", "", "-a, --aa"},
		{"short first placeholder", "a, aa", "foo", "-a, --aa <foo>"},
		{"long first placeholder", "aa, a", "foo", "-a, --aa <foo>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			got := flagPrefixer(tt.flags, tt.placeholder)
			g.Should(be.Equal(got, tt.want))
		})
	}
}

func TestCreateArgsSection(t *testing.T) {
	tests := []struct {
		name    string
		taskCfg string
		want    string
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

	for _, tt := range tests {
		taskName := "someTaskName"
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			cfgText := fmt.Sprintf("tasks: { %s: { args: {%s} } }", taskName, tt.taskCfg)
			cfg, err := runner.Parse([]byte(cfgText))
			g.NoError(err)

			got := createArgsSection(cfg.Tasks[taskName])
			g.Should(be.Equal(got, tt.want))
		})
	}
}
