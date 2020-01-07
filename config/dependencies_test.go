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
			createOption(
				withOptionName("foo"),
			),
		},
		[]int{},
		[]int{},
	},
	{
		"fake dependencies",
		[]Option{
			createOption(
				withOptionName("foo"),
				withOptionDependency("fake"),
			),
		},
		[]Option{},
		[]int{0},
		[]int{},
	},
	{
		"multiple dependencies per option",
		[]Option{
			createOption(
				withOptionName("foo"),
				withOptionDependency("one"),
				withOptionDependency("two"),
				withOptionDependency("three"),
			),
		},
		[]Option{
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
		[]int{0},
		[]int{0, 1, 2},
	},
	{
		"only task dependencies",
		[]Option{
			createOption(
				withOptionName("foo"),
			),
			createOption(
				withOptionName("bar"),
				withOptionDependency("foo"),
			),
		},
		[]Option{},
		[]int{0, 1},
		[]int{},
	},
	{
		"overridden global dependencies",
		[]Option{
			createOption(
				withOptionName("foo"),
			),
			createOption(
				withOptionName("bar"),
				withOptionDependency("foo"),
			),
		},
		[]Option{
			createOption(
				withOptionName("foo"),
			),
		},
		[]int{0, 1},
		[]int{},
	},
	{
		"when dependencies",
		[]Option{
			createOption(
				withOptionName("foo"),
			),
			createOption(
				withOptionName("bar"),
				withOptionWhenDependency("foo"),
			),
		},
		[]Option{},
		[]int{0, 1},
		[]int{},
	},
	{
		"task requires global",
		[]Option{
			createOption(
				withOptionName("bar"),
				withOptionDependency("foo"),
			),
		},
		[]Option{
			createOption(
				withOptionName("foo"),
			),
		},
		[]int{0},
		[]int{0},
	},
	{
		"global requires task (false positive)",
		[]Option{
			createOption(
				withOptionName("foo"),
			),
		},
		[]Option{
			createOption(
				withOptionName("bar"),
				withOptionDependency("foo"),
			),
		},
		[]int{0},
		[]int{},
	},
	{
		"nested depdendencies",
		[]Option{
			createOption(
				withOptionName("foo"),
				withOptionDependency("bar"),
			),
			createOption(
				withOptionName("bar"),
				withOptionDependency("baz"),
			),
		},
		[]Option{
			createOption(
				withOptionName("baz"),
				withOptionDependency("qux"),
			),
			createOption(
				withOptionName("qux"),
			),
		},
		[]int{0, 1},
		[]int{0, 1},
	},
	{
		"nested depdendencies with ignored globals",
		[]Option{
			createOption(
				withOptionName("foo"),
				withOptionDependency("bar"),
			),
			createOption(
				withOptionName("bar"),
				withOptionDependency("baz"),
			),
		},
		[]Option{
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
