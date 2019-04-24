package config

import (
	"io/ioutil"
	"path/filepath"

	"github.com/rliebz/tusk/ui"
)

// Metadata contains global configuration settings.
type Metadata struct {
	CfgText             []byte
	Directory           string
	InstallCompletion   string
	UninstallCompletion string
	PrintHelp           bool
	PrintVersion        bool
	Verbosity           ui.VerbosityLevel
}

// Set sets the metadata based on options.
func (m *Metadata) Set(o OptGetter) error {
	var err error

	fullPath := o.String("file")
	if fullPath != "" {
		if m.CfgText, err = ioutil.ReadFile(fullPath); err != nil {
			return err
		}
	} else {
		var found bool
		fullPath, found, err = searchForFile()
		if err != nil {
			return err
		}

		if found {
			if m.CfgText, err = ioutil.ReadFile(fullPath); err != nil {
				return err
			}
		}
	}

	m.InstallCompletion = o.String("install-completion")
	m.UninstallCompletion = o.String("uninstall-completion")
	m.Directory = filepath.Dir(fullPath)
	m.PrintHelp = o.Bool("help")
	m.PrintVersion = o.Bool("version")
	m.Verbosity = getVerbosity(o)
	return nil
}

// OptGetter pulls various options based on a name.
// These options will generally come from the command line.
type OptGetter interface {
	Bool(string) bool
	String(string) string
}

func getVerbosity(c OptGetter) ui.VerbosityLevel {
	switch {
	case c.Bool("silent"):
		return ui.VerbosityLevelSilent
	case c.Bool("quiet"):
		return ui.VerbosityLevelQuiet
	case c.Bool("verbose"):
		return ui.VerbosityLevelVerbose
	default:
		return ui.VerbosityLevelNormal
	}
}
