package runner

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
)

func TestFindAllOptions(t *testing.T) {
	tests := []struct {
		name            string
		taskOptions     []*Option
		cfgOptions      []*Option
		wantTaskIndices []int
		wantCfgIndices  []int
	}{
		{
			name: "no dependencies",
			cfgOptions: []*Option{
				createOption(
					withOptionName("foo"),
				),
			},
		},
		{
			name: "fake dependencies",
			taskOptions: []*Option{
				createOption(
					withOptionName("foo"),
					withOptionDependency("fake"),
				),
			},
			wantTaskIndices: []int{0},
		},
		{
			name: "multiple dependencies per option",
			taskOptions: []*Option{
				createOption(
					withOptionName("foo"),
					withOptionDependency("one"),
					withOptionDependency("two"),
					withOptionDependency("three"),
				),
			},
			cfgOptions: []*Option{
				createOption(
					withOptionName("one"),
				),
				createOption(
					withOptionName("two"),
				),
				createOption(
					withOptionName("three"),
				),
				createOption(
					withOptionName("wrong"),
				),
			},
			wantTaskIndices: []int{0},
			wantCfgIndices:  []int{0, 1, 2},
		},
		{
			name: "only task dependencies",
			taskOptions: []*Option{
				createOption(
					withOptionName("foo"),
				),
				createOption(
					withOptionName("bar"),
					withOptionDependency("foo"),
				),
			},
			wantTaskIndices: []int{0, 1},
		},
		{
			name: "overridden global dependencies",
			taskOptions: []*Option{
				createOption(
					withOptionName("foo"),
				),
				createOption(
					withOptionName("bar"),
					withOptionDependency("foo"),
				),
			},
			cfgOptions: []*Option{
				createOption(
					withOptionName("foo"),
				),
			},
			wantTaskIndices: []int{0, 1},
		},
		{
			name: "when dependencies",
			taskOptions: []*Option{
				createOption(
					withOptionName("foo"),
				),
				createOption(
					withOptionName("bar"),
					withOptionWhenDependency("foo"),
				),
			},
			wantTaskIndices: []int{0, 1},
		},
		{
			name: "task requires global",
			taskOptions: []*Option{
				createOption(
					withOptionName("bar"),
					withOptionDependency("foo"),
				),
			},
			cfgOptions: []*Option{
				createOption(
					withOptionName("foo"),
				),
			},
			wantTaskIndices: []int{0},
			wantCfgIndices:  []int{0},
		},
		{
			name: "global requires task (false positive)",
			taskOptions: []*Option{
				createOption(
					withOptionName("foo"),
				),
			},
			cfgOptions: []*Option{
				createOption(
					withOptionName("bar"),
					withOptionDependency("foo"),
				),
			},
			wantTaskIndices: []int{0},
		},
		{
			name: "nested depdendencies",
			taskOptions: []*Option{
				createOption(
					withOptionName("foo"),
					withOptionDependency("bar"),
				),
				createOption(
					withOptionName("bar"),
					withOptionDependency("baz"),
				),
			},
			cfgOptions: []*Option{
				createOption(
					withOptionName("baz"),
					withOptionDependency("qux"),
				),
				createOption(
					withOptionName("qux"),
				),
			},
			wantTaskIndices: []int{0, 1},
			wantCfgIndices:  []int{0, 1},
		},
		{
			name: "nested depdendencies with ignored globals",
			taskOptions: []*Option{
				createOption(
					withOptionName("foo"),
					withOptionDependency("bar"),
				),
				createOption(
					withOptionName("bar"),
					withOptionDependency("baz"),
				),
			},
			cfgOptions: []*Option{
				createOption(
					withOptionName("qux"),
				),
				createOption(
					withOptionName("skiptwo"),
				),
				createOption(
					withOptionName("baz"),
					withOptionDependency("qux"),
				),
				createOption(
					withOptionName("skipone"),
					withOptionDependency("skiptwo"),
				),
			},
			wantTaskIndices: []int{0, 1},
			wantCfgIndices:  []int{0, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			var tsk Task
			for _, opt := range tt.taskOptions {
				tsk.Options = append(tsk.Options, opt)
			}

			var cfg Config
			for _, opt := range tt.cfgOptions {
				cfg.Options = append(cfg.Options, opt)
			}

			got, err := FindAllOptions(&tsk, &cfg)
			g.NoError(err)

			var want []*Option
			for _, i := range tt.wantTaskIndices {
				want = append(want, tt.taskOptions[i])
			}
			for _, i := range tt.wantCfgIndices {
				want = append(want, tt.cfgOptions[i])
			}

			assertOptionsEqualUnordered(t, want, got)
		})
	}
}

func assertOptionsEqualUnordered(t *testing.T, a, b []*Option) {
	t.Helper()

	g := ghost.New(t)

	if !g.Should(be.SliceLen(b, len(a))) {
		return
	}

	bMap := make(map[*Option]any)
	for _, val := range b {
		bMap[val] = struct{}{}
	}

	for _, item := range a {
		_, ok := bMap[item]
		if !g.Check(ok) {
			t.Log("missing item:", item)
		}
	}
}
