package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

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
	Interpreter         []string
	Directory           string
	InstallCompletion   string
	UninstallCompletion string
	PrintHelp           bool
	PrintVersion        bool
	Logger              *ui.Logger
}

// Set sets the metadata based on options.
func (m *Metadata) Set(o OptGetter) error {
	var fullPath string
	var err error
	fullPath, m.CfgText, err = getConfigFile(o)
	if err != nil {
		return err
	}

	m.Interpreter, err = getInterpreter(m.CfgText)
	if err != nil {
		return err
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

func getConfigFile(o OptGetter) (fullPath string, cfgText []byte, _ error) {
	fullPath = o.String("file")

	if fullPath == "" {
		var err error
		fullPath, err = searchForFile()
		if err != nil || fullPath == "" {
			return "", nil, err
		}
	}

	cfgText, err := os.ReadFile(fullPath)
	if err != nil {
		return "", nil, fmt.Errorf("reading config file %q: %w", fullPath, err)
	}

	return fullPath, cfgText, nil
}

// getInterpreter attempts to determine the interpreter by reading the config
// file. This should occur before full config parsing, as it may influence the
// interpretation of option and arg resolutions.
//
// If no interpreter is specified, nil will be returned.
func getInterpreter(cfgText []byte) ([]string, error) {
	var cfg struct {
		Interpreter string `yaml:"interpreter"`
	}

	if err := yaml.Unmarshal(cfgText, &cfg); err != nil {
		return nil, err
	}

	if cfg.Interpreter == "" {
		return nil, nil
	}

	return strings.Fields(cfg.Interpreter), nil
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
