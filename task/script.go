package task

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/pkg/errors"

	"gitlab.com/rliebz/tusk/ui"
)

// Script is a single script within a task
type Script struct {
	When struct {
		Exists []string `yaml:",omitempty"`
		OS     []string `yaml:",omitempty"`
		Test   []string `yaml:",omitempty"`
	} `yaml:",omitempty"`
	Run []string
}

// Execute validates the When conditions and executes a Script.
func (script Script) Execute() error {

	for _, f := range script.When.Exists {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			fmt.Printf("File not found: %s\n", f)
			return nil
		}
	}

	for _, os := range script.When.OS {
		if runtime.GOOS != os {
			fmt.Printf("Unexpected Architecture: %s\n", os)
			return nil
		}
	}

	for _, test := range script.When.Test {
		if err := testCommand(test); err != nil {
			fmt.Printf("Test failed: %s\n", test)
			return nil
		}
	}

	for _, command := range script.Run {
		err := execCommand(command)
		if err != nil {
			return err
		}
	}

	return nil
}

func testCommand(test string) error {
	args := strings.Fields(test)
	_, err := exec.Command("test", args...).Output() // nolint: gas
	return err
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
