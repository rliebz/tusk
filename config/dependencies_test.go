package config

import (
	"testing"

	"github.com/rliebz/tusk/config/option"
)

func TestAddNestedDependencies_none(t *testing.T) {
	actual, err := addNestedDependencies([]*option.Option{}, []*option.Option{})
	if err != nil {
		t.Fatalf(`addNestedDependencies([], []): unexpected error: %s`, err)
	}

	if len(actual) != 0 {
		t.Errorf(
			`addNestedDependencies([], []): expected empty slice, got: %v`,
			actual,
		)
	}
}

func TestAddNestedDependencies_combines(t *testing.T) {
	dependencies := []*option.Option{
		{Name: "One"},
	}

	nested := []*option.Option{
		{Name: "Two"},
		{Name: "Three"},
	}
	actual, err := addNestedDependencies(dependencies, nested)
	if err != nil {
		t.Fatalf(`addNestedDependencies(): unexpected error: %s`, err)
	}

	if len(actual) != 3 {
		t.Errorf(`addNestedDependencies(): expected 3 items, got: %+v`, actual)
	}
}

func TestAddNestedDependencies_disallows_redefines(t *testing.T) {
	dependencies := []*option.Option{
		{Name: "One"},
	}

	nested := []*option.Option{
		{Name: "One"},
	}
	if _, err := addNestedDependencies(dependencies, nested); err == nil {
		t.Fatal(
			`addNestedDependencies(): expected error for redefining option, got nil`,
		)
	}
}

func TestAddNestedDependencies_allows_duplicates(t *testing.T) {
	duplicate := &option.Option{Name: "Dupl"}
	dependencies := []*option.Option{
		{Name: "One"},
		duplicate,
	}

	nested := []*option.Option{
		{Name: "Two"},
		{Name: "Three"},
		duplicate,
	}
	actual, err := addNestedDependencies(dependencies, nested)
	if err != nil {
		t.Fatalf(`addNestedDependencies(): unexpected error: %s`, err)
	}

	if len(actual) != 4 {
		t.Errorf(`addNestedDependencies(): expected 4 items, got: %+v`, actual)
	}
}
