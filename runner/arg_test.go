package runner

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	"github.com/rliebz/tusk/marshal"
	yaml "gopkg.in/yaml.v2"
)

func TestEvaluate(t *testing.T) {
	g := ghost.New(t)

	want := "foo"
	arg := Arg{
		Passable: Passable{
			Passed: want,
		},
	}

	got, err := arg.Evaluate()
	g.NoError(err)

	g.Should(be.Equal(want, got))
}

func TestEvaluate_specified(t *testing.T) {
	g := ghost.New(t)

	want := "foo"
	arg := Arg{
		Passable: Passable{
			Passed:        want,
			ValuesAllowed: marshal.StringList{"wrong", want, "other"},
		},
	}

	got, err := arg.Evaluate()
	g.NoError(err)

	g.Should(be.Equal(want, got))
}

func TestEvaluate_unspecified(t *testing.T) {
	g := ghost.New(t)

	passed := "foo"
	arg := Arg{
		Passable: Passable{
			Name:          "my-arg",
			Passed:        passed,
			ValuesAllowed: marshal.StringList{"wrong", "other"},
		},
	}

	_, err := arg.Evaluate()
	g.Should(be.ErrorEqual(`value "foo" for argument "my-arg" must be one of [wrong other]`, err))
}

func TestEvaluate_nil(t *testing.T) {
	g := ghost.New(t)

	var arg *Arg
	_, err := arg.Evaluate()
	g.Should(be.ErrorEqual("nil argument evaluated", err))
}

func TestGetArgsWithOrder(t *testing.T) {
	g := ghost.New(t)

	ms := yaml.MapSlice{
		{
			Key: "foo",
			Value: &Arg{
				Passable: Passable{
					Usage: "first usage",
				},
			},
		},
		{
			Key: "bar",
			Value: &Arg{
				Passable: Passable{
					Usage: "other usage",
				},
			},
		},
	}

	args, err := getArgsWithOrder(ms)
	g.NoError(err)

	g.Must(be.SliceLen(2, args))

	g.Should(be.Equal("foo", args[0].Name))
	g.Should(be.Equal("first usage", args[0].Usage))
	g.Should(be.Equal("bar", args[1].Name))
	g.Should(be.Equal("other usage", args[1].Usage))
}

func TestGetArgsWithOrder_invalid(t *testing.T) {
	g := ghost.New(t)

	ms := yaml.MapSlice{
		{Key: "foo", Value: "not an arg"},
	}

	_, err := getArgsWithOrder(ms)
	g.Should(be.ErrorContaining("cannot unmarshal !!str `not an arg` into runner.Arg", err))
}
