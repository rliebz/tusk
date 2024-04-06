package runner

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	yaml "gopkg.in/yaml.v2"
)

func TestSubTask_UnmarshalYAML(t *testing.T) {
	g := ghost.New(t)

	var st1 SubTask
	err := yaml.UnmarshalStrict([]byte(`name: example`), &st1)
	g.NoError(err)

	var st2 SubTask
	err = yaml.UnmarshalStrict([]byte(`example`), &st2)
	g.NoError(err)

	g.Should(be.DeepEqual(st1, st2))
	g.Should(be.DeepEqual(st1, SubTask{Name: "example"}))
}
