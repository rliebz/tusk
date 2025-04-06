package appcli

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	"github.com/urfave/cli"

	"github.com/rliebz/tusk/runner"
)

func TestCreateCLIFlag_undefined(t *testing.T) {
	g := ghost.New(t)

	opt := &runner.Option{
		Passable: runner.Passable{
			Type: "wrong",
		},
	}

	flag, err := createCLIFlag(opt)
	g.Should(be.ErrorEqual(err, `unsupported flag type "wrong"`))

	g.Should(be.Nil(flag))
}

func TestAddFlag_no_duplicates(t *testing.T) {
	g := ghost.New(t)

	command := &cli.Command{}

	opt := &runner.Option{
		Passable: runner.Passable{
			Name: "foo",
		},
		Short: "f",
	}

	err := addFlag(command, opt)
	g.NoError(err)

	err = addFlag(command, opt)
	g.NoError(err)

	g.Should(be.SliceLen(command.Flags, 1))
}
