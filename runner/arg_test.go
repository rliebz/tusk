package runner

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	yaml "gopkg.in/yaml.v2"

	"github.com/rliebz/tusk/marshal"
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

	g.Should(be.Equal(got, want))
}

func TestEvaluate_specified(t *testing.T) {
	g := ghost.New(t)

	want := "foo"
	arg := Arg{
		Passable: Passable{
			Passed:        want,
			ValuesAllowed: marshal.Slice[string]{"wrong", want, "other"},
		},
	}

	got, err := arg.Evaluate()
	g.NoError(err)

	g.Should(be.Equal(got, want))
}

func TestEvaluate_unspecified(t *testing.T) {
	g := ghost.New(t)

	passed := "foo"
	arg := Arg{
		Passable: Passable{
			Name:          "my-arg",
			Passed:        passed,
			ValuesAllowed: marshal.Slice[string]{"wrong", "other"},
		},
	}

	_, err := arg.Evaluate()
	g.Should(be.ErrorEqual(err, `value "foo" for argument "my-arg" must be one of [wrong, other]`))
}

func TestEvaluate_nil(t *testing.T) {
	g := ghost.New(t)

	var arg *Arg
	_, err := arg.Evaluate()
	g.Should(be.ErrorEqual(err, "nil argument evaluated"))
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

	g.Must(be.SliceLen(args, 2))

	g.Should(be.Equal(args[0].Name, "foo"))
	g.Should(be.Equal(args[0].Usage, "first usage"))
	g.Should(be.Equal(args[1].Name, "bar"))
	g.Should(be.Equal(args[1].Usage, "other usage"))
}

func TestGetArgsWithOrder_invalid(t *testing.T) {
	g := ghost.New(t)

	ms := yaml.MapSlice{
		{Key: "foo", Value: "not an arg"},
	}

	_, err := getArgsWithOrder(ms)
	g.Should(be.ErrorContaining(err, "cannot unmarshal !!str `not an arg` into runner.Arg"))
}
