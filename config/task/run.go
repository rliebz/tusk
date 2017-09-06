package task

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/pkg/errors"
	"github.com/rliebz/tusk/appyaml"
	"github.com/rliebz/tusk/ui"
)

// run defines a a single runnable script within a task.
type run struct {
	When    *appyaml.When      `yaml:",omitempty"`
	Command appyaml.StringList `yaml:",omitempty"`
	Task    appyaml.StringList `yaml:",omitempty"`
}

// UnmarshalYAML allows plain strings to represent a run struct. The value of
// the string is used as the Command field.
func (r *run) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var command string
	if err := unmarshal(&command); err == nil {
		*r = run{Command: appyaml.StringList{command}}
		return nil
	}

	type runType run // Use new type to avoid recursion
	var runItem *runType
	if err := unmarshal(&runItem); err == nil {
		*r = *(*run)(runItem)
		return nil
	}

	return errors.New("could not parse run item")
}

type runList []*run

// UnmarshalYAML allows single items to be used as lists.
func (rl *runList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var runItem *run
	if err := unmarshal(&runItem); err == nil {
		*rl = runList{runItem}
		return nil
	}

	var runSlice []*run
	if err := unmarshal(&runSlice); err == nil {
		*rl = runSlice
		return nil
	}

	return errors.New("could not parse runlist")
}

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

func execCommand(command string) error {
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
