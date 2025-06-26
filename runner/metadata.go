package runner

import (
	"cmp"
	"fmt"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/rliebz/tusk/ui"
)

// Metadata contains global configuration settings.
type Metadata struct {
	CfgPath     string
	CfgText     []byte
	Interpreter []string
	Logger      *ui.Logger

	InstallCompletion   string
	UninstallCompletion string
	PrintHelp           bool
	PrintVersion        bool
	CleanCache          bool
	CleanProjectCache   bool
	CleanTaskCache      string
}

// Set sets the metadata based on options.
func (m *Metadata) Set(o OptGetter) error {
	var err error
	cfgPath, cfgText, err := getConfigFile(o)
	if err != nil {
		return err
	}

	interpreter, err := getInterpreter(cfgText)
	if err != nil {
		return err
	}

	m.CfgPath, m.CfgText = cfgPath, cfgText
	m.Interpreter = interpreter
	m.InstallCompletion = o.String("install-completion")
	m.UninstallCompletion = o.String("uninstall-completion")
	m.PrintHelp = o.Bool("help")
	m.PrintVersion = o.Bool("version")
	m.CleanCache = o.Bool("clean-cache")
	m.CleanProjectCache = o.Bool("clean-project-cache")
	m.CleanTaskCache = o.String("clean-task-cache")

	m.Logger = cmp.Or(m.Logger, ui.New(ui.Config{}))
	m.Logger.SetLevel(getVerbosity(o))
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

func getVerbosity(c OptGetter) ui.Level {
	switch {
	case c.Bool("silent"):
		return ui.LevelSilent
	case c.Bool("quiet"):
		return ui.LevelQuiet
	case c.Bool("verbose"):
		return ui.LevelVerbose
	default:
		return ui.LevelNormal
	}
}
