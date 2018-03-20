package appcli

import (
	"testing"

	"github.com/urfave/cli"
)

var isCompletingTests = []struct {
	flags    []cli.Flag
	arg      string
	expected bool
}{
	{[]cli.Flag{}, "foo", false},
	{[]cli.Flag{}, "-f", false},
	{[]cli.Flag{}, "--foo", false},
	{[]cli.Flag{cli.BoolFlag{Name: "f, foo"}}, "-f", false},
	{[]cli.Flag{cli.BoolFlag{Name: "f, foo"}}, "--foo", false},
	{[]cli.Flag{cli.BoolTFlag{Name: "f, foo"}}, "-f", false},
	{[]cli.Flag{cli.BoolTFlag{Name: "f, foo"}}, "--foo", false},
	{[]cli.Flag{cli.StringFlag{Name: "f, foo"}}, "-f", true},
	{[]cli.Flag{cli.StringFlag{Name: "f, foo"}}, "--foo", true},
	{[]cli.Flag{cli.StringFlag{Name: "f, foo"}}, "--f", false},
	{[]cli.Flag{cli.StringFlag{Name: "b, bar"}}, "-f", false},
	{[]cli.Flag{cli.StringFlag{Name: "b, bar"}}, "--foo", false},
	{[]cli.Flag{cli.StringFlag{Name: "f, foo"}, cli.StringFlag{Name: "b, bar"}}, "-f", true},
	{[]cli.Flag{cli.StringFlag{Name: "f, foo"}, cli.StringFlag{Name: "b, bar"}}, "--foo", true},
}

func TestIsCompletingArg(t *testing.T) {
	for _, tt := range isCompletingTests {
		actual := isCompletingArg(tt.flags, tt.arg)
		if tt.expected != actual {
			t.Errorf(
				"isCompletingArg(%#v, %s) => %t, want %t",
				tt.flags, tt.arg, actual, tt.expected,
			)
		}
	}
}
