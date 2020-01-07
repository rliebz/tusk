package config

import (
	"testing"
)

var findalloptionstests = []struct {
	desc                string
	taskOptions         []Option
	cfgOptions          []Option
	expectedTaskIndices []int
	expectedCfgIndices  []int
}{
	{
		"no dependencies",
		[]Option{},
		[]Option{
			Create(
				WithName("foo"),
			),
		},
		[]int{},
		[]int{},
	},
	{
		"fake dependencies",
		[]Option{
			Create(
				WithName("foo"),
				WithDependency("fake"),
			),
		},
		[]Option{},
		[]int{0},
		[]int{},
	},
	{
		"multiple dependencies per option",
		[]Option{
			Create(
				WithName("foo"),
				WithDependency("one"),
				WithDependency("two"),
				WithDependency("three"),
			),
		},
		[]Option{
			Create(
				WithName("one"),
			),
			Create(
				WithName("two"),
			),
			Create(
				WithName("three"),
			),
			Create(
				WithName("wrong"),
			),
		},
		[]int{0},
		[]int{0, 1, 2},
	},
	{
		"only task dependencies",
		[]Option{
			Create(
				WithName("foo"),
			),
			Create(
				WithName("bar"),
				WithDependency("foo"),
			),
		},
		[]Option{},
		[]int{0, 1},
		[]int{},
	},
	{
		"overridden global dependencies",
		[]Option{
			Create(
				WithName("foo"),
			),
			Create(
				WithName("bar"),
				WithDependency("foo"),
			),
		},
		[]Option{
			Create(
				WithName("foo"),
			),
		},
		[]int{0, 1},
		[]int{},
	},
	{
		"when dependencies",
		[]Option{
			Create(
				WithName("foo"),
			),
			Create(
				WithName("bar"),
				WithWhenDependency("foo"),
			),
		},
		[]Option{},
		[]int{0, 1},
		[]int{},
	},
	{
		"task requires global",
		[]Option{
			Create(
				WithName("bar"),
				WithDependency("foo"),
			),
		},
		[]Option{
			Create(
				WithName("foo"),
			),
		},
		[]int{0},
		[]int{0},
	},
	{
		"global requires task (false positive)",
		[]Option{
			Create(
				WithName("foo"),
			),
		},
		[]Option{
			Create(
				WithName("bar"),
				WithDependency("foo"),
			),
		},
		[]int{0},
		[]int{},
	},
	{
		"nested depdendencies",
		[]Option{
			Create(
				WithName("foo"),
				WithDependency("bar"),
			),
			Create(
				WithName("bar"),
				WithDependency("baz"),
			),
		},
		[]Option{
			Create(
				WithName("baz"),
				WithDependency("qux"),
			),
			Create(
				WithName("qux"),
			),
		},
		[]int{0, 1},
		[]int{0, 1},
	},
	{
		"nested depdendencies with ignored globals",
		[]Option{
			Create(
				WithName("foo"),
				WithDependency("bar"),
			),
			Create(
				WithName("bar"),
				WithDependency("baz"),
			),
		},
		[]Option{
			Create(
				WithName("qux"),
			),
			Create(
				WithName("skiptwo"),
			),
			Create(
				WithName("baz"),
				WithDependency("qux"),
			),
			Create(
				WithName("skipone"),
				WithDependency("skiptwo"),
			),
		},
		[]int{0, 1},
		[]int{0, 2},
	},
}

func TestFindAllOptions(t *testing.T) {
	for _, tt := range findalloptionstests {
		tsk := Task{}
		for i := range tt.taskOptions {
			tsk.Options = append(tsk.Options, &tt.taskOptions[i])
		}

		cfg := Config{}
		for i := range tt.cfgOptions {
			cfg.Options = append(cfg.Options, &tt.cfgOptions[i])
		}

		actual, err := FindAllOptions(&tsk, &cfg)
		if err != nil {
			t.Errorf(
				"FindAllOptions() for %s: unexpected error: %s",
				tt.desc, err,
			)
			continue
		}

		var expected []*Option
		for _, i := range tt.expectedTaskIndices {
			expected = append(expected, &tt.taskOptions[i])
		}
		for _, i := range tt.expectedCfgIndices {
			expected = append(expected, &tt.cfgOptions[i])
		}

		assertOptionsEqualUnordered(t, tt.desc, expected, actual)
	}
}

func assertOptionsEqualUnordered(t *testing.T, desc string, a, b []*Option) {
	t.Helper()

	if len(a) != len(b) {
		t.Errorf(
			"options for %s: expected %d options, actual %d",
			desc, len(a), len(b),
		)
		return
	}

	bMap := make(map[*Option]interface{})
	for _, val := range b {
		bMap[val] = struct{}{}
	}

	for _, item := range a {
		if _, ok := bMap[item]; !ok {
			t.Errorf(
				"options for %s: expected item %s not in actual list",
				desc, item.Name,
			)
			continue
		}
	}
}
