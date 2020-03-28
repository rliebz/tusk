package runner

import (
	"io/ioutil"
	"path/filepath"

	"github.com/rliebz/tusk/ui"
)

// NewMetadata creates a metadata struct with a default logger.
func NewMetadata() *Metadata {
	return &Metadata{
		Logger: ui.New(),
	}
}

// Metadata contains global configuration settings.
//
// Metadata should be instantiated using NewMetadata.
type Metadata struct {
	CfgText             []byte
	Directory           string
	InstallCompletion   string
	UninstallCompletion string
	PrintHelp           bool
	PrintVersion        bool
	Logger              *ui.Logger
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
	if m.Logger == nil {
		m.Logger = ui.New()
	}
	m.Logger.Verbosity = getVerbosity(o)
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
