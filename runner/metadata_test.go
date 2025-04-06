package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	"gotest.tools/v3/fs"

	"github.com/rliebz/tusk/ui"
)

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

			var meta Metadata
			err = meta.Set(opts)
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

			var meta Metadata
			err := meta.Set(opts)
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
