package option

import (
	"reflect"
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
		valueWithList: valueWithList{
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
		valueWithList: valueWithList{
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
	usage := "foo usage"
	ms := yaml.MapSlice{
		{Key: name, Value: &Arg{Usage: usage}},
		{Key: "bar", Value: &Arg{Usage: "bar usage"}},
	}

	args, ordered, err := GetArgsWithOrder(ms)
	if err != nil {
		t.Fatalf("GetArgsWithOrder(ms) => unexpected error: %v", err)
	}

	if len(ms) != len(args) || len(ms) != len(ordered) {
		t.Fatalf(
			"GetArgsWithOrder(ms) => want %d items, got %d in map and %d in slice",
			len(ms), len(args), len(ordered),
		)
	}

	arg, ok := args[name]
	if !ok {
		t.Fatalf("GetArgsWithOrder(ms) => item %q is not in map", name)
	}

	if name != arg.Name {
		t.Errorf(
			"GetArgsWithOrder(ms) => want arg.Name %q, got %q",
			name, arg.Name,
		)
	}

	if usage != arg.Usage {
		t.Errorf(
			"GetArgsWithOrder(ms) => want arg.Usage %q, got %q",
			usage, arg.Usage,
		)
	}

	expectedOrder := []string{"foo", "bar"}
	if !reflect.DeepEqual(expectedOrder, ordered) {
		t.Errorf(
			"GetArgsWithOrder(ms) => want ordered %v, got %v",
			expectedOrder, ordered,
		)
	}

}

func TestGetArgsWithOrder_null_arg(t *testing.T) {
	ms := yaml.MapSlice{
		{Key: "foo", Value: nil},
	}

	_, _, err := GetArgsWithOrder(ms)
	if err == nil {
		t.Error("GetArgsWithOrder() => expected error for null argument")
	}
}

func TestGetArgsWithOrder_invalid(t *testing.T) {
	ms := yaml.MapSlice{
		{Key: "foo", Value: "not an arg"},
	}

	_, _, err := GetArgsWithOrder(ms)
	if err == nil {
		t.Error("GetArgsWithOrder() => expected yaml parsing error")
	}
}
