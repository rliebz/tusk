package run

import (
	"bytes"
	"log"
	"testing"

	"github.com/pkg/errors"
	"github.com/rliebz/tusk/ui"
)

func TestExecCommand_output(t *testing.T) {

	command := `
echo one
>&2 echo two
>&2 echo three
echo four
`

	bufExpected := new(bytes.Buffer)
	ui.Stderr = log.New(bufExpected, "", 0)
	ui.PrintCommand(command)
	ui.PrintCommandOutput("one")
	ui.PrintCommandOutput("two")
	ui.PrintCommandOutput("three")
	ui.PrintCommandOutput("four")
	expected := bufExpected.String()

	bufActual := new(bytes.Buffer)
	ui.Stderr = log.New(bufActual, "", 0)
	if err := ExecCommand(command); err != nil {
		t.Fatalf(`execCommand("%s"): unexpected err: %s`, command, err)
	}
	actual := bufActual.String()

	if expected != actual {
		t.Fatalf(
			"execCommand(\"%s\"):\nexpected output:\n`%s`\nactual output:\n`%s`",
			command, expected, actual,
		)
	}
}

func TestExecCommand_error(t *testing.T) {

	command := "exit 1"

	bufExpected := new(bytes.Buffer)
	errExpected := errors.New("exit status 1")
	ui.Stderr = log.New(bufExpected, "", 0)
	ui.PrintCommand(command)
	ui.PrintCommandError(errExpected)

	expected := bufExpected.String()

	bufActual := new(bytes.Buffer)
	ui.Stderr = log.New(bufActual, "", 0)
	if err := ExecCommand(command); err.Error() != errExpected.Error() {
		t.Fatalf(`execCommand("%s"): expected error "%s", actual "%s"`,
			command, errExpected, err,
		)
	}
	actual := bufActual.String()

	if expected != actual {
		t.Fatalf(
			"execCommand(\"%s\"):\nexpected output:\n`%s`\nactual output:\n`%s`",
			command, expected, actual,
		)
	}
}
