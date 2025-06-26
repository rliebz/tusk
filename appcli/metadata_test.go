package appcli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	"gotest.tools/v3/fs"

	"github.com/rliebz/tusk/ui"
)

func TestNewMetadata_defaults(t *testing.T) {
	g := ghost.New(t)

	args := []string{"tusk"}

	meta, err := NewMetadata(ui.Noop(), args)
	g.NoError(err)

	// The project's tuskfile should be found in the project root.
	wd, err := os.Getwd()
	g.NoError(err)

	g.Should(be.Equal(meta.CfgPath, filepath.Join(filepath.Dir(wd), "tusk.yml")))
	g.Should(be.Equal(meta.Logger.Level(), ui.LevelNormal))
	g.Should(be.False(meta.PrintVersion))
}

func TestNewMetadata_file(t *testing.T) {
	g := ghost.New(t)

	cfgPath := "testdata/example.yml"
	args := []string{"tusk", "--file", cfgPath}

	meta, err := NewMetadata(ui.Noop(), args)
	g.NoError(err)

	g.Should(be.Equal(meta.CfgPath, cfgPath))

	cfgText, err := os.ReadFile(cfgPath)
	g.NoError(err)

	g.Should(be.Equal(string(meta.CfgText), string(cfgText)))
}

func TestNewMetadata_fileNoExist(t *testing.T) {
	g := ghost.New(t)

	_, err := NewMetadata(ui.Noop(), []string{"tusk", "--file", "fakefile.yml"})
	if !g.Should(be.True(errors.Is(err, os.ErrNotExist))) {
		t.Log(err)
	}
}

func TestNewMetadata_version(t *testing.T) {
	g := ghost.New(t)

	meta, err := NewMetadata(ui.Noop(), []string{"tusk", "--version"})
	g.NoError(err)

	g.Should(be.True(meta.PrintVersion))
}

func TestNewMetadata_log_level(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want ui.Level
	}{
		{
			"normal",
			[]string{"tusk"},
			ui.LevelNormal,
		},
		{
			"silent",
			[]string{"tusk", "--silent"},
			ui.LevelSilent,
		},
		{
			"quiet",
			[]string{"tusk", "--quiet"},
			ui.LevelQuiet,
		},
		{
			"verbose",
			[]string{"tusk", "--verbose"},
			ui.LevelVerbose,
		},
		{
			"quiet verbose",
			[]string{"tusk", "--quiet", "--verbose"},
			ui.LevelQuiet,
		},
		{
			"silent quiet",
			[]string{"tusk", "--silent", "--quiet"},
			ui.LevelSilent,
		},
		{
			"silent verbose",
			[]string{"tusk", "--silent", "--verbose"},
			ui.LevelSilent,
		},
		{
			"silent quiet verbose",
			[]string{"tusk", "--silent", "--quiet", "--verbose"},
			ui.LevelSilent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			meta, err := NewMetadata(ui.Noop(), tt.args)
			g.NoError(err)

			g.Should(be.Equal(meta.Logger.Level(), tt.want))
		})
	}
}

// mockOptGetter returns opts from maps.
type mockOptGetter struct {
	bools   map[string]bool
	strings map[string]string
}

func (m mockOptGetter) String(v string) string {
	if m.strings != nil {
		return m.strings[v]
	}

	return ""
}

func (m mockOptGetter) Bool(v string) bool {
	if m.bools != nil {
		return m.bools[v]
	}

	return false
}

func TestMetadata_Set(t *testing.T) {
	dirEmpty := fs.NewDir(t, "empty-dir")

	dirFullContents := `value: yaml config found in dir`
	dirFull := fs.NewDir(t, "full-dir", fs.WithFile("tusk.yml", dirFullContents))

	cfgFileContents := `value: yaml config passed from --file`
	cfgFile := fs.NewFile(t, "", fs.WithContent(cfgFileContents))

	normal := ui.New(ui.Config{Verbosity: ui.LevelNormal})
	silent := ui.New(ui.Config{Verbosity: ui.LevelSilent})
	quiet := ui.New(ui.Config{Verbosity: ui.LevelQuiet})
	verbose := ui.New(ui.Config{Verbosity: ui.LevelVerbose})

	tests := []struct {
		name    string
		bools   map[string]bool
		strings map[string]string
		meta    Metadata
		wd      string
	}{
		{
			name: "defaults",
			meta: Metadata{
				Logger: normal,
			},
		},
		{
			name: "passed-config-file",
			strings: map[string]string{
				"file": cfgFile.Path(),
			},
			meta: Metadata{
				CfgPath: cfgFile.Path(),
				CfgText: []byte(cfgFileContents),
				Logger:  normal,
			},
		},
		{
			name: "found-config-file",
			meta: Metadata{
				CfgPath: filepath.Join(dirFull.Path(), "tusk.yml"),
				CfgText: []byte(dirFullContents),
				Logger:  normal,
			},
			wd: dirFull.Path(),
		},
		{
			name: "passed-overwrites-found",
			strings: map[string]string{
				"file": cfgFile.Path(),
			},
			meta: Metadata{
				CfgPath: cfgFile.Path(),
				CfgText: []byte(cfgFileContents),
				Logger:  normal,
			},
			wd: dirFull.Path(),
		},
		{
			name: "install-completion",
			strings: map[string]string{
				"install-completion": "zsh",
			},
			meta: Metadata{
				InstallCompletion: "zsh",
				Logger:            normal,
			},
		},
		{
			name: "uninstall-completion",
			strings: map[string]string{
				"uninstall-completion": "zsh",
			},
			meta: Metadata{
				UninstallCompletion: "zsh",
				Logger:              normal,
			},
		},
		{
			name: "print-help",
			bools: map[string]bool{
				"help": true,
			},
			meta: Metadata{
				PrintHelp: true,
				Logger:    normal,
			},
		},
		{
			name: "print-version",
			bools: map[string]bool{
				"version": true,
			},
			meta: Metadata{
				PrintVersion: true,
				Logger:       normal,
			},
		},
		{
			name: "verbosity-silent",
			bools: map[string]bool{
				"silent": true,
			},
			meta: Metadata{
				Logger: silent,
			},
		},
		{
			name: "verbosity-quiet",
			bools: map[string]bool{
				"quiet": true,
			},
			meta: Metadata{
				Logger: quiet,
			},
		},
		{
			name: "verbosity-verbose",
			bools: map[string]bool{
				"verbose": true,
			},
			meta: Metadata{
				Logger: verbose,
			},
		},
		{
			name: "verbosity-prefers-silence",
			bools: map[string]bool{
				"silent":  true,
				"quiet":   true,
				"verbose": true,
			},
			meta: Metadata{
				Logger: silent,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			stashEnv(t)

			cwd, err := os.Getwd()
			g.NoError(err)

			// default to no config for irrelevant tests
			if tt.wd == "" {
				tt.wd = dirEmpty.Path()
			}

			err = os.Chdir(tt.wd)
			g.NoError(err)
			t.Cleanup(func() {
				os.Chdir(cwd) //nolint: errcheck
			})

			opts := mockOptGetter{
				bools:   tt.bools,
				strings: tt.strings,
			}

			meta := Metadata{Logger: ui.New(ui.Config{})}
			err = meta.set(opts)
			g.NoError(err)

			// evaluate symlinks to avoid noise
			meta.CfgPath, err = filepath.EvalSymlinks(meta.CfgPath)
			g.NoError(err)

			tt.meta.CfgPath, err = filepath.EvalSymlinks(tt.meta.CfgPath)
			g.NoError(err)

			g.Should(be.DeepEqual(meta, tt.meta))
		})
	}
}

func TestMetadata_Set_interpreter(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		want    []string
		wantErr string
	}{
		{
			name: "defaults",
		},
		{
			name:   "executable",
			config: `interpreter: bash`,
			want:   []string{"bash"},
		},
		{
			name:   "executable with arg",
			config: `interpreter: /usr/bin/env node -e`,
			want:   []string{"/usr/bin/env", "node", "-e"},
		},
		{
			name:   "invalid yaml",
			config: "🥔",
			wantErr: `yaml: unmarshal errors:
  line 1: cannot unmarshal !!str ` + "`🥔`" +
				` into struct { Interpreter string "yaml:\"interpreter\"" }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			cfgFile := fs.NewFile(t, "", fs.WithContent(tt.config))
			opts := mockOptGetter{
				strings: map[string]string{"file": cfgFile.Path()},
			}

			meta := Metadata{Logger: ui.Noop()}
			err := meta.set(opts)
			if tt.wantErr != "" {
				g.Should(be.ErrorEqual(err, tt.wantErr))
				return
			}
			g.NoError(err)

			g.Should(be.DeepEqual(meta.Interpreter, tt.want))
		})
	}
}

func stashEnv(t testing.TB) {
	t.Helper()

	environ := os.Environ()

	t.Cleanup(func() {
		for _, val := range environ {
			parts := strings.Split(val, "=")
			os.Setenv(parts[0], parts[1]) //nolint:errcheck,usetesting
		}
	})

	os.Clearenv()
}
