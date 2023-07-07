package runner

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
)

func TestContext_TaskNames(t *testing.T) {
	g := ghost.New(t)

	var ctx Context
	g.Should(be.SliceLen(0, ctx.TaskNames()))

	ctx.PushTask(&Task{Name: "foo"})
	ctx.PushTask(&Task{Name: "bar"})

	g.Should(be.DeepEqual([]string{"foo", "bar"}, ctx.TaskNames()))
}
