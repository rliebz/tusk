package appcli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rliebz/tusk/ui"
)

const zshInstallDir = "/usr/local/share/zsh/site-functions"

// InstallCompletions installs command line completions for a given shell.
func InstallCompletions(shell string) error {
	switch shell {
	case "bash":
		return installBashCompletion()
	case "zsh":
		return installZshCompletion(zshInstallDir)
	default:
		return fmt.Errorf("tab completion for %q is not supported", shell)
	}
}

// UninstallCompletions uninstalls command line completions for a given shell.
func UninstallCompletions(shell string) error {
	switch shell {
	case "zsh":
		return uninstallZshCompletion(zshInstallDir)
	default:
		return fmt.Errorf("tab completion for %q is not supported", shell)
	}
}

// TODO: Write the real command
const bashCommand = `echo "Hello, world"`

var bashRCFiles = []string{".bashrc", ".bash_profile", ".profile"}

func installBashCompletion() error {
	rcfile, err := getBashRCFile()
	if err != nil {
		return err
	}

	return appendIfAbsent(rcfile, bashCommand)
}

func getBashRCFile() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	for _, rcfile := range bashRCFiles {
		path := filepath.Join(homedir, rcfile)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return "", err
		}

		return path, nil
	}

	return filepath.Join(homedir, ".bashrc"), nil
}

func appendIfAbsent(path, text string) error {
	// nolint: gosec
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close() // nolint: errcheck

	scanner := bufio.NewScanner(f)

	isEmpty := true
	for scanner.Scan() {
		isEmpty = false
		if scanner.Text() == text {
			return nil
		}
	}
	if serr := scanner.Err(); serr != nil {
		return serr
	}

	if !isEmpty {
		text = "\n" + text
	}

	_, err = fmt.Fprintln(f, text)
	return err
}

func installZshCompletion(dir string) error {
	// nolint: gosec
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	target := filepath.Join(dir, "_tusk")
	if err := ioutil.WriteFile(target, []byte(rawZshCompletion), 0644); err != nil {
		return err
	}

	ui.Info("zsh completions successfully installed", target)
	return nil
}

func uninstallZshCompletion(dir string) error {
	err := os.Remove(filepath.Join(dir, "_tusk"))
	if !os.IsNotExist(err) {
		return err
	}

	return nil
}
