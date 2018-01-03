package optiontest

import (
	"fmt"
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestWithName(t *testing.T) {
	expected := "foo"
	o := Create(WithName(expected))
	if expected != o.Name {
		t.Errorf(
			`expected name: "%s", actual: "%s"`,
			expected, o.Name,
		)
	}
}

func TestWithDependency(t *testing.T) {
	a := "foo"
	b := "bar"

	expectedA := fmt.Sprintf("${%s}", a)
	expectedB := fmt.Sprintf("${%s}", b)

	o := Create(
		WithDependency(a),
		WithDependency(b),
	)

	actual, err := yaml.Marshal(o)
	if err != nil {
		t.Fatalf("unexpected error marshalling option: %s", err)
	}

	if !strings.Contains(string(actual), expectedA) {
		t.Errorf("option does not contain string: %s", expectedA)
	}

	if !strings.Contains(string(actual), expectedB) {
		t.Errorf("option does not contain string: %s", expectedB)
	}
}

func TestWithWhenDependency(t *testing.T) {
	a := "foo"
	b := "bar"

	foundA := false
	foundB := false

	o := Create(
		WithWhenDependency(a),
		WithWhenDependency(b),
	)

	for _, value := range o.DefaultValues {
		for key := range value.When.Equal {
			if key == a {
				foundA = true
			} else if key == b {
				foundB = true
			}
		}
	}

	if !foundA {
		t.Errorf("option does not contain when value: %s", a)
	}

	if !foundB {
		t.Errorf("option does not contain when value: %s", b)
	}
}
