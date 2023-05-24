package appcli

import (
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/tusk/runner"
	"github.com/urfave/cli"
)

func TestCreateCLIFlag_undefined(t *testing.T) {
	opt := &runner.Option{
		Passable: runner.Passable{
			Type: "wrong",
		},
	}

	flag, err := createCLIFlag(opt)
	if err == nil {
		t.Fatalf("flag was wrongly created: %#v", flag)
	}
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

	g.Should(ghost.Equal(1, len(command.Flags)))
}
