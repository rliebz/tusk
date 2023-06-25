package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/tusk/ui"
	"gotest.tools/v3/fs"
)

// mockOptGetter returns opts from maps
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

	normal := ui.New()
	silent := ui.New()
	silent.Verbosity = ui.VerbosityLevelSilent
	quiet := ui.New()
	quiet.Verbosity = ui.VerbosityLevelQuiet
	verbose := ui.New()
	verbose.Verbosity = ui.VerbosityLevelVerbose

	tests := []struct {
		name    string
		bools   map[string]bool
		strings map[string]string
		meta    Metadata
		wd      string
	}{
		{
			"defaults",
			nil,
			nil,
			Metadata{
				Directory: ".",
				Logger:    normal,
			},
			"",
		},
		{
			"passed-config-file",
			nil,
			map[string]string{
				"file": cfgFile.Path(),
			},
			Metadata{
				CfgText:   []byte(cfgFileContents),
				Directory: filepath.Dir(cfgFile.Path()),
				Logger:    normal,
			},
			"",
		},
		{
			"found-config-file",
			nil,
			nil,
			Metadata{
				CfgText:   []byte(dirFullContents),
				Directory: dirFull.Path(),
				Logger:    normal,
			},
			dirFull.Path(),
		},
		{
			"passed-overwrites-found",
			nil,
			map[string]string{
				"file": cfgFile.Path(),
			},
			Metadata{
				CfgText:   []byte(cfgFileContents),
				Directory: filepath.Dir(cfgFile.Path()),
				Logger:    normal,
			},
			dirFull.Path(),
		},
		{
			"install-completion",
			nil,
			map[string]string{
				"install-completion": "zsh",
			},
			Metadata{
				Directory:         ".",
				InstallCompletion: "zsh",
				Logger:            normal,
			},
			"",
		},
		{
			"uninstall-completion",
			nil,
			map[string]string{
				"uninstall-completion": "zsh",
			},
			Metadata{
				Directory:           ".",
				UninstallCompletion: "zsh",
				Logger:              normal,
			},
			"",
		},
		{
			"print-help",
			map[string]bool{
				"help": true,
			},
			nil,
			Metadata{
				Directory: ".",
				PrintHelp: true,
				Logger:    normal,
			},
			"",
		},
		{
			"print-version",
			map[string]bool{
				"version": true,
			},
			nil,
			Metadata{
				Directory:    ".",
				PrintVersion: true,
				Logger:       normal,
			},
			"",
		},
		{
			"verbosity-silent",
			map[string]bool{
				"silent": true,
			},
			nil,
			Metadata{
				Directory: ".",
				Logger:    silent,
			},
			"",
		},
		{
			"verbosity-quiet",
			map[string]bool{
				"quiet": true,
			},
			nil,
			Metadata{
				Directory: ".",
				Logger:    quiet,
			},
			"",
		},
		{
			"verbosity-verbose",
			map[string]bool{
				"verbose": true,
			},
			nil,
			Metadata{
				Directory: ".",
				Logger:    verbose,
			},
			"",
		},
		{
			"verbosity-prefers-silence",
			map[string]bool{
				"silent":  true,
				"quiet":   true,
				"verbose": true,
			},
			nil,
			Metadata{
				Directory: ".",
				Logger:    silent,
			},
			"",
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

			var meta Metadata
			err = meta.Set(opts)
			g.NoError(err)

			// evaluate symlinks to avoid noise
			meta.Directory, err = filepath.EvalSymlinks(meta.Directory)
			g.NoError(err)

			tt.meta.Directory, err = filepath.EvalSymlinks(tt.meta.Directory)
			g.NoError(err)

			g.Should(ghost.DeepEqual(tt.meta, meta))
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

			var meta Metadata
			err := meta.Set(opts)
			if tt.wantErr != "" {
				g.Should(ghost.ErrorEqual(err, tt.wantErr))
				return
			}
			g.NoError(err)

			g.Should(ghost.DeepEqual(tt.want, meta.Interpreter))
		})
	}
}

func stashEnv(t testing.TB) {
	t.Helper()

	environ := os.Environ()

	t.Cleanup(func() {
		for _, val := range environ {
			parts := strings.Split(val, "=")
			os.Setenv(parts[0], parts[1]) //nolint: errcheck
		}
	})

	os.Clearenv()
}
