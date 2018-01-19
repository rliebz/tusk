package config

import (
	"testing"

	"github.com/rliebz/tusk/config/option"
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
			option.Create(
				option.WithName("foo"),
			),
		},
		[]int{},
		[]int{},
	},
	{
		"fake dependencies",
		[]option.Option{
			option.Create(
				option.WithName("foo"),
				option.WithDependency("fake"),
			),
		},
		[]option.Option{},
		[]int{0},
		[]int{},
	},
	{
		"multiple dependencies per option",
		[]option.Option{
			option.Create(
				option.WithName("foo"),
				option.WithDependency("one"),
				option.WithDependency("two"),
				option.WithDependency("three"),
			),
		},
		[]option.Option{
			option.Create(
				option.WithName("one"),
			),
			option.Create(
				option.WithName("two"),
			),
			option.Create(
				option.WithName("three"),
			),
			option.Create(
				option.WithName("wrong"),
			),
		},
		[]int{0},
		[]int{0, 1, 2},
	},
	{
		"only task dependencies",
		[]option.Option{
			option.Create(
				option.WithName("foo"),
			),
			option.Create(
				option.WithName("bar"),
				option.WithDependency("foo"),
			),
		},
		[]option.Option{},
		[]int{0, 1},
		[]int{},
	},
	{
		"overridden global dependencies",
		[]option.Option{
			option.Create(
				option.WithName("foo"),
			),
			option.Create(
				option.WithName("bar"),
				option.WithDependency("foo"),
			),
		},
		[]option.Option{
			option.Create(
				option.WithName("foo"),
			),
		},
		[]int{0, 1},
		[]int{},
	},
	{
		"when dependencies",
		[]option.Option{
			option.Create(
				option.WithName("foo"),
			),
			option.Create(
				option.WithName("bar"),
				option.WithWhenDependency("foo"),
			),
		},
		[]option.Option{},
		[]int{0, 1},
		[]int{},
	},
	{
		"task requires global",
		[]option.Option{
			option.Create(
				option.WithName("bar"),
				option.WithDependency("foo"),
			),
		},
		[]option.Option{
			option.Create(
				option.WithName("foo"),
			),
		},
		[]int{0},
		[]int{0},
	},
	{
		"global requires task (false positive)",
		[]option.Option{
			option.Create(
				option.WithName("foo"),
			),
		},
		[]option.Option{
			option.Create(
				option.WithName("bar"),
				option.WithDependency("foo"),
			),
		},
		[]int{0},
		[]int{},
	},
	{
		"nested depdendencies",
		[]option.Option{
			option.Create(
				option.WithName("foo"),
				option.WithDependency("bar"),
			),
			option.Create(
				option.WithName("bar"),
				option.WithDependency("baz"),
			),
		},
		[]option.Option{
			option.Create(
				option.WithName("baz"),
				option.WithDependency("qux"),
			),
			option.Create(
				option.WithName("qux"),
			),
		},
		[]int{0, 1},
		[]int{0, 1},
	},
	{
		"nested depdendencies with ignored globals",
		[]option.Option{
			option.Create(
				option.WithName("foo"),
				option.WithDependency("bar"),
			),
			option.Create(
				option.WithName("bar"),
				option.WithDependency("baz"),
			),
		},
		[]option.Option{
			option.Create(
				option.WithName("qux"),
			),
			option.Create(
				option.WithName("skiptwo"),
			),
			option.Create(
				option.WithName("baz"),
				option.WithDependency("qux"),
			),
			option.Create(
				option.WithName("skipone"),
				option.WithDependency("skiptwo"),
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
	t.Helper()

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
