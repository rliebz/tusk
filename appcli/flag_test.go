package appcli

import (
	"testing"

	"github.com/rliebz/tusk/config/option"
	"github.com/urfave/cli"
)

func TestCreateCLIFlag_undefined(t *testing.T) {
	opt := &option.Option{
		Type: "wrong",
	}

	flag, err := createCLIFlag(opt)
	if err == nil {
		t.Fatalf("flag was wrongly created: %#v", flag)
	}
}

func TestAddFlag_no_duplicates(t *testing.T) {

	command := &cli.Command{}

	opt := &option.Option{
		Name:  "foo",
		Short: "f",
	}

	if err := addFlag(command, opt); err != nil {
		t.Fatalf(`addFlag(): unexpected err: %s`, err)
	}

	if err := addFlag(command, opt); err != nil {
		t.Fatalf(`addFlag(): unexpected err: %s`, err)
	}

	if len(command.Flags) != 1 {
		t.Errorf(
			`addFlag() twice with same flag: expected %d flags, actual %d`,
			2, len(command.Flags),
		)
	}

}
