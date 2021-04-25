package appcli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	// Embed completion scripts.
	_ "embed"

	"github.com/rliebz/tusk/runner"
	"github.com/rliebz/tusk/ui"
)

//go:embed completion/tusk-completion.bash
var rawBashCompletion string

//go:embed completion/tusk.fish
var rawFishCompletion string

//go:embed completion/_tusk
var rawZshCompletion string

const (
	bashCompletionFile = "tusk-completion.bash"
	fishCompletionFile = "tusk.fish"
	zshCompletionFile  = "_tusk"
	zshInstallDir      = "/usr/local/share/zsh/site-functions"
)

var bashRCFiles = []string{".bashrc", ".bash_profile", ".profile"}

// InstallCompletion installs command line tab completion for a given shell.
func InstallCompletion(meta *runner.Metadata) error {
	shell := meta.InstallCompletion
	switch shell {
	case "bash":
		return installBashCompletion(meta.Logger)
	case "fish":
		return installFishCompletion(meta.Logger)
	case "zsh":
		return installZshCompletion(meta.Logger, zshInstallDir)
	default:
		return fmt.Errorf("tab completion for %q is not supported", shell)
	}
}

// UninstallCompletion uninstalls command line tab completion for a given shell.
func UninstallCompletion(meta *runner.Metadata) error {
	shell := meta.UninstallCompletion
	switch shell {
	case "bash":
		return uninstallBashCompletion()
	case "fish":
		return uninstallFishCompletion()
	case "zsh":
		return uninstallZshCompletion(zshInstallDir)
	default:
		return fmt.Errorf("tab completion for %q is not supported", shell)
	}
}

func installBashCompletion(logger *ui.Logger) error {
	dir, err := getDataDir()
	if err != nil {
		return err
	}

	err = installFileInDir(logger, dir, bashCompletionFile, []byte(rawBashCompletion))
	if err != nil {
		return err
	}

	rcfile, err := getBashRCFile()
	if err != nil {
		return err
	}

	slashPath := filepath.ToSlash(filepath.Join(dir, bashCompletionFile))
	command := fmt.Sprintf("source %q", slashPath)
	return appendIfAbsent(rcfile, command)
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

	prependNewline := false
	for scanner.Scan() {
		switch scanner.Text() {
		case text:
			return nil
		case "":
			prependNewline = false
		default:
			prependNewline = true
		}
	}
	if serr := scanner.Err(); serr != nil {
		return serr
	}

	if prependNewline {
		text = "\n" + text
	}

	_, err = fmt.Fprintln(f, text)
	return err
}

func uninstallBashCompletion() error {
	dir, err := getDataDir()
	if err != nil {
		return err
	}

	err = uninstallFileFromDir(dir, bashCompletionFile)
	if err != nil {
		return err
	}

	rcfile, err := getBashRCFile()
	if err != nil {
		return err
	}

	re := regexp.MustCompile(fmt.Sprintf(`source ".*/%s"`, bashCompletionFile))
	return removeLineInFile(rcfile, re)
}

func removeLineInFile(path string, re *regexp.Regexp) error {
	rf, err := os.OpenFile(path, os.O_RDONLY, 0644) // nolint: gosec
	if err != nil {
		return err
	}
	defer rf.Close() // nolint: errcheck

	wf, err := ioutil.TempFile("", ".profile.tusk.bkp")
	if err != nil {
		return err
	}
	defer wf.Close() // nolint: errcheck

	scanner := bufio.NewScanner(rf)

	buf := ""
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case re.MatchString(line):
			continue
		case line == "":
			buf += "\n"
			continue
		}

		_, err := fmt.Fprintln(wf, buf+line)
		if err != nil {
			return err
		}

		buf = ""
	}
	if serr := scanner.Err(); serr != nil {
		return serr
	}

	rf.Close() // nolint: errcheck
	wf.Close() // nolint: errcheck
	return os.Rename(wf.Name(), path)
}

func installFishCompletion(logger *ui.Logger) error {
	dir, err := getFishCompletionsDir()
	if err != nil {
		return err
	}

	return installFileInDir(logger, dir, fishCompletionFile, []byte(rawFishCompletion))
}

func uninstallFishCompletion() error {
	dir, err := getFishCompletionsDir()
	if err != nil {
		return err
	}

	return uninstallFileFromDir(dir, fishCompletionFile)
}

func getDataDir() (string, error) {
	if xdgHome := os.Getenv("XDG_DATA_HOME"); xdgHome != "" {
		return filepath.Join(xdgHome, "tusk"), nil
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homedir, ".local", "share", "tusk"), nil
}

// getFishCompletionsDir gets the directory to place completions in, adhering
// to the XDG base directory.
func getFishCompletionsDir() (string, error) {
	if xdgHome := os.Getenv("XDG_CONFIG_HOME"); xdgHome != "" {
		return filepath.Join(xdgHome, "fish", "completions"), nil
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homedir, ".config", "fish", "completions"), nil
}

func installZshCompletion(logger *ui.Logger, dir string) error {
	return installFileInDir(logger, dir, zshCompletionFile, []byte(rawZshCompletion))
}

func installFileInDir(logger *ui.Logger, dir, file string, content []byte) error {
	// nolint: gosec
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	target := filepath.Join(dir, file)

	// nolint: gosec
	if err := ioutil.WriteFile(target, content, 0644); err != nil {
		return err
	}

	logger.Info("Tab completion successfully installed", target)
	logger.Info("You may need to restart your shell for completion to take effect")
	return nil
}

func uninstallZshCompletion(dir string) error {
	return uninstallFileFromDir(dir, zshCompletionFile)
}

func uninstallFileFromDir(dir, file string) error {
	err := os.Remove(filepath.Join(dir, file))
	if !os.IsNotExist(err) {
		return err
	}

	return nil
}
