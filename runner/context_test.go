package runner

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
)

func TestContext_TaskNames(t *testing.T) {
	g := ghost.New(t)

	var ctx0 Context
	g.Should(be.SliceLen(ctx0.TaskNames(), 0))

	ctx := ctx0
	ctx = ctx.WithTask(&Task{Name: "foo"})
	ctx = ctx.WithTask(&Task{Name: "bar"})

	g.Should(be.DeepEqual(ctx.TaskNames(), []string{"foo", "bar"}))
	g.Should(be.SliceLen(ctx0.TaskNames(), 0))
}
