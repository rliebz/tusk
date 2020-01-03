package option

import (
	"testing"

	"github.com/rliebz/tusk/config/marshal"
	yaml "gopkg.in/yaml.v2"
)

func TestEvaluate(t *testing.T) {
	expected := "foo"
	arg := Arg{Passed: expected}

	actual, err := arg.Evaluate()
	if err != nil {
		t.Fatalf("Arg.Evaluate() => unexpected error: %v", err)
	}

	if expected != actual {
		t.Errorf("Arg.Evaluate() => want %q, got %q", expected, actual)
	}
}

func TestEvaluate_specified(t *testing.T) {
	expected := "foo"
	arg := Arg{
		Passed: expected,
		ValueWithList: ValueWithList{
			ValuesAllowed: marshal.StringList{"wrong", expected, "other"},
		},
	}

	actual, err := arg.Evaluate()
	if err != nil {
		t.Fatalf("Arg.Evaluate() => unexpected error: %v", err)
	}

	if expected != actual {
		t.Errorf("Arg.Evaluate() => want %q, got %q", expected, actual)
	}
}

func TestEvaluate_unspecified(t *testing.T) {
	passed := "foo"
	arg := Arg{
		Passed: passed,
		ValueWithList: ValueWithList{
			ValuesAllowed: marshal.StringList{"wrong", "other"},
		},
	}

	if _, err := arg.Evaluate(); err == nil {
		t.Fatal("Arg.Evaluate() => want error for nil argument, got nil")
	}
}

func TestEvaluate_nil(t *testing.T) {
	var arg *Arg
	if _, err := arg.Evaluate(); err == nil {
		t.Fatal("Arg.Evaluate() => want error for nil argument, got nil")
	}
}

// nolint: dupl
func TestGetArgsWithOrder(t *testing.T) {
	name := "foo"
	usage := "use me"
	ms := yaml.MapSlice{
		{Key: name, Value: &Arg{Usage: usage}},
		{Key: "bar", Value: &Arg{Usage: "other usage"}},
	}

	args, err := getArgsWithOrder(ms)
	if err != nil {
		t.Fatalf("GetArgsWithOrder(ms) => unexpected error: %v", err)
	}

	if len(ms) != len(args) {
		t.Fatalf(
			"GetArgsWithOrder(ms) => want %d items, got %d",
			len(ms), len(args),
		)
	}

	opt := args[0]

	if name != opt.Name {
		t.Errorf(
			"GetArgsWithOrder(ms) => want opt.Name %q, got %q",
			name, opt.Name,
		)
	}

	if usage != opt.Usage {
		t.Errorf(
			"GetArgsWithOrder(ms) => want arg.Usage %q, got %q",
			usage, opt.Usage,
		)
	}

	if args[1].Name != "bar" {
		t.Errorf("GetArgsWithOrder(ms) => want 2nd arg %q, got %q", "bar", args[1].Name)
	}
}

func TestGetArgsWithOrder_invalid(t *testing.T) {
	ms := yaml.MapSlice{
		{Key: "foo", Value: "not an arg"},
	}

	_, err := getArgsWithOrder(ms)
	if err == nil {
		t.Error("GetArgsWithOrder() => expected yaml parsing error")
	}
}
