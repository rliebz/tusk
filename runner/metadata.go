package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rliebz/tusk/ui"
	yaml "gopkg.in/yaml.v2"
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
	fullPath := o.String("file")

	if fullPath == "" {
		var err error
		fullPath, err = searchForFile()
		if err != nil {
			return err
		}
	}

	if fullPath != "" {
		var err error
		m.CfgText, err = os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("reading config file %q: %w", fullPath, err)
		}
	}

	interpreter, err := getInterpreter(m.CfgText)
	if err != nil {
		return err
	}

	m.Interpreter = interpreter
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

// getInterpreter attempts to determine the interpreter from reading the env
// var and the config file, in that order. If no interpreter is specified, "sh"
// is used.
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
