package config

import (
	"testing"

	"github.com/rliebz/tusk/config/option"
	"github.com/rliebz/tusk/config/optiontest"
	"github.com/rliebz/tusk/config/task"
)

var findalloptionstests = []struct {
	desc                string
	taskOptions         []option.Option
	cfgOptions          []option.Option
	expectedTaskIndices []int
	expectedCfgIndices  []int
}{
	{
		"no dependencies",
		[]option.Option{},
		[]option.Option{
			optiontest.Create(
				optiontest.WithName("foo"),
			),
		},
		[]int{},
		[]int{},
	},
	{
		"fake dependencies",
		[]option.Option{
			optiontest.Create(
				optiontest.WithName("foo"),
				optiontest.WithDependency("fake"),
			),
		},
		[]option.Option{},
		[]int{0},
		[]int{},
	},
	{
		"only task dependencies",
		[]option.Option{
			optiontest.Create(
				optiontest.WithName("foo"),
			),
			optiontest.Create(
				optiontest.WithName("bar"),
				optiontest.WithDependency("foo"),
			),
		},
		[]option.Option{},
		[]int{0, 1},
		[]int{},
	},
	{
		"overridden global dependencies",
		[]option.Option{
			optiontest.Create(
				optiontest.WithName("foo"),
			),
			optiontest.Create(
				optiontest.WithName("bar"),
				optiontest.WithDependency("foo"),
			),
		},
		[]option.Option{
			optiontest.Create(
				optiontest.WithName("foo"),
			),
		},
		[]int{0, 1},
		[]int{},
	},
	{
		"when dependencies",
		[]option.Option{
			optiontest.Create(
				optiontest.WithName("foo"),
			),
			optiontest.Create(
				optiontest.WithName("bar"),
				optiontest.WithWhenDependency("foo"),
			),
		},
		[]option.Option{},
		[]int{0, 1},
		[]int{},
	},
	{
		"task requires global",
		[]option.Option{
			optiontest.Create(
				optiontest.WithName("bar"),
				optiontest.WithDependency("foo"),
			),
		},
		[]option.Option{
			optiontest.Create(
				optiontest.WithName("foo"),
			),
		},
		[]int{0},
		[]int{0},
	},
	{
		"global requires task (false positive)",
		[]option.Option{
			optiontest.Create(
				optiontest.WithName("foo"),
			),
		},
		[]option.Option{
			optiontest.Create(
				optiontest.WithName("bar"),
				optiontest.WithDependency("foo"),
			),
		},
		[]int{0},
		[]int{},
	},
	{
		"nested depdendencies",
		[]option.Option{
			optiontest.Create(
				optiontest.WithName("foo"),
				optiontest.WithDependency("bar"),
			),
			optiontest.Create(
				optiontest.WithName("bar"),
				optiontest.WithDependency("baz"),
			),
		},
		[]option.Option{
			optiontest.Create(
				optiontest.WithName("baz"),
				optiontest.WithDependency("qux"),
			),
			optiontest.Create(
				optiontest.WithName("qux"),
			),
		},
		[]int{0, 1},
		[]int{0, 1},
	},
	{
		"nested depdendencies with ignored globals",
		[]option.Option{
			optiontest.Create(
				optiontest.WithName("foo"),
				optiontest.WithDependency("bar"),
			),
			optiontest.Create(
				optiontest.WithName("bar"),
				optiontest.WithDependency("baz"),
			),
		},
		[]option.Option{
			optiontest.Create(
				optiontest.WithName("qux"),
			),
			optiontest.Create(
				optiontest.WithName("skiptwo"),
			),
			optiontest.Create(
				optiontest.WithName("baz"),
				optiontest.WithDependency("qux"),
			),
			optiontest.Create(
				optiontest.WithName("skipone"),
				optiontest.WithDependency("skiptwo"),
			),
		},
		[]int{0, 1},
		[]int{0, 2},
	},
}

func TestFindAllOptions(t *testing.T) {

	for _, tt := range findalloptionstests {

		tsk := task.Task{
			Options: map[string]*option.Option{},
		}
		for i, o := range tt.taskOptions {
			tsk.Options[o.Name] = &tt.taskOptions[i]
		}

		cfg := Config{
			Options: map[string]*option.Option{},
		}
		for i, o := range tt.cfgOptions {
			cfg.Options[o.Name] = &tt.cfgOptions[i]
		}

		actual, err := FindAllOptions(&tsk, &cfg)
		if err != nil {
			t.Errorf(
				"FindAllOptions() for %s: unexpected error: %s",
				tt.desc, err,
			)
			continue
		}

		var expected []*option.Option
		for _, i := range tt.expectedTaskIndices {
			expected = append(expected, &tt.taskOptions[i])
		}
		for _, i := range tt.expectedCfgIndices {
			expected = append(expected, &tt.cfgOptions[i])
		}

		assertOptionsEqualUnordered(t, tt.desc, expected, actual)
	}
}

func assertOptionsEqualUnordered(t *testing.T, desc string, a, b []*option.Option) {
	if len(a) != len(b) {
		t.Errorf(
			"options for %s: expected %d options, actual %d",
			desc, len(a), len(b),
		)
		return
	}

	bMap := make(map[*option.Option]interface{})
	for _, val := range b {
		bMap[val] = struct{}{}
	}

	for _, item := range a {
		if _, ok := bMap[item]; !ok {
			t.Errorf(
				"expected item %s not in actual list",
				item.Name,
			)
			continue
		}
	}
}
