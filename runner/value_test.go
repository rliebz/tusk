package runner

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	yaml "gopkg.in/yaml.v2"
)

func TestValue_UnmarshalYAML(t *testing.T) {
	g := ghost.New(t)

	var v1 Value
	err := yaml.UnmarshalStrict([]byte(`value: example`), &v1)
	g.NoError(err)

	var v2 Value
	err = yaml.UnmarshalStrict([]byte(`example`), &v2)
	g.NoError(err)

	g.Should(be.DeepEqual(v1, v2))
	g.Should(be.Equal(v1.Value, "example"))
}

func TestValue_UnmarshalYAML_value_and_command(t *testing.T) {
	g := ghost.New(t)

	var v Value
	err := yaml.UnmarshalStrict([]byte(`{value: "example", command: "echo hello"}`), &v)
	g.Should(be.ErrorEqual(err, "value (example) and command (echo hello) are both defined"))
}

func TestValueList_UnmarshalYAML(t *testing.T) {
	g := ghost.New(t)

	var v1 ValueList
	err := yaml.UnmarshalStrict([]byte(`example`), &v1)
	g.NoError(err)

	var v2 ValueList
	err = yaml.UnmarshalStrict([]byte(`[example]`), &v2)
	g.NoError(err)

	g.Should(be.DeepEqual(v1, v2))
	g.Should(be.DeepEqual(v1, ValueList{{Value: "example"}}))
}
