package run

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/pkg/errors"
	"github.com/rliebz/tusk/ui"
)

// waitWriter wraps a writer with a wait group.
// This ca ensure there are no additional writes pending.
type waitWriter struct {
	writer    io.Writer
	waitGroup *sync.WaitGroup
}

func (w waitWriter) Write(p []byte) (int, error) {
	w.waitGroup.Add(len(p))
	return w.writer.Write(p)
}

// ExecCommand executes a shell command.
func ExecCommand(command string) error {
	if ui.Quiet {
		return execCommandQuiet(command)
	}

	return execCommandWithStyle(command)
}

func execCommandQuiet(command string) error {
	ui.PrintCommand(command)

	cmd := exec.Command("sh", "-c", command) // nolint: gas
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func execCommandWithStyle(command string) error {
	ui.PrintCommand(command)

	cmd := exec.Command("sh", "-c", command) // nolint: gas

	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}
	defer closeFile(pw)

	wg := sync.WaitGroup{}
	ww := waitWriter{pw, &wg}

	cmd.Stdin = os.Stdin
	cmd.Stdout = ww
	cmd.Stderr = ww

	scanner := bufio.NewScanner(pr)
	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			ui.PrintCommandOutput(text)
			for i := 0; i <= len(text); i++ {
				wg.Done()
			}
		}
	}()

	if err := cmd.Run(); err != nil {
		ui.PrintCommandError(err)
		return err
	}

	wg.Wait()
	return nil
}

func closeFile(file *os.File) {
	if err := file.Close(); err != nil {
		ui.Error(errors.Wrap(err, "Failed to close file"))
	}
}
