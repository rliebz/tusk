package appcli

import (
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
