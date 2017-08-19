package task

import (
	"bufio"
	"os"
	"os/exec"

	"github.com/pkg/errors"

	"gitlab.com/rliebz/tusk/appyaml"
	"gitlab.com/rliebz/tusk/ui"
)

// Run defines a a single runnable script within a task.
type Run struct {
	When    appyaml.When `yaml:",omitempty"`
	Command appyaml.StringList
}

// Execute validates the When conditions and executes a Run script.
func (run Run) Execute() error {

	if err := run.When.Validate(); err != nil {
		for _, command := range run.Command.Values {
			ui.PrintCommandSkipped(command, err.Error())
		}
		return nil
	}

	for _, command := range run.Command.Values {
		err := execCommand(command)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: Handle errors
func execCommand(command string) error {
	ui.PrintCommand(command)

	cmd := exec.Command("sh", "-c", command) // nolint: gas

	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}
	defer closeFile(pr)
	defer closeFile(pw)

	// TODO: Is it possible to keep the output ordered and separate?
	cmd.Stdout = pw
	cmd.Stderr = pw

	scanner := bufio.NewScanner(pr)
	go func() {
		for scanner.Scan() {
			ui.PrintCommandOutput(scanner.Text())
		}
	}()

	if err := cmd.Run(); err != nil {
		ui.PrintCommandError(err)
		return err
	}

	return nil
}

func closeFile(file *os.File) {
	if err := file.Close(); err != nil {
		ui.Error(errors.Wrap(err, "Failed to close file"))
	}
}
