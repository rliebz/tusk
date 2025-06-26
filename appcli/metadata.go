package appcli

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
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

// NewMetadata returns a metadata object based on global options passed.
func NewMetadata(logger *ui.Logger, args []string) (*Metadata, error) {
	app := newSilentApp()
	metadata := Metadata{Logger: logger}

	var err error
	app.Action = func(c *cli.Context) error {
		// To prevent app from exiting, app.Action must return nil on error.
		// The enclosing function will still return the error.
		err = metadata.set(c)
		return nil
	}
	if runErr := populateMetadata(app, args); runErr != nil {
		return nil, runErr
	}
	return &metadata, err
}

// optGetter pulls various options based on a name.
// These options will generally come from the command line.
type optGetter interface {
	Bool(string) bool
	String(string) string
}

// set sets the metadata based on options.
func (m *Metadata) set(o optGetter) error {
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
	m.Logger.SetLevel(getLogLevel(o))
	return nil
}

// populateMetadata runs the app to populate the metadata struct.
func populateMetadata(app *cli.App, args []string) error {
	args = removeCompletionArg(args)

	if err := app.Run(args); err != nil {
		// Ignore flags without arguments during metadata creation
		if isFlagArgumentError(err) {
			return app.Run(args[:len(args)-1])
		}

		return err
	}

	return nil
}

func getConfigFile(o optGetter) (fullPath string, cfgText []byte, _ error) {
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

func getLogLevel(c optGetter) ui.Level {
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

func isFlagArgumentError(err error) bool {
	return strings.HasPrefix(err.Error(), "flag needs an argument")
}
