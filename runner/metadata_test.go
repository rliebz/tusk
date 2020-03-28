package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
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
	defer dirEmpty.Remove()

	dirFullContents := `yaml config found in dir`
	dirFull := fs.NewDir(t, "full-dir", fs.WithFile("tusk.yml", dirFullContents))
	defer dirFull.Remove()

	cfgFileContents := `yaml config passed from --file`
	cfgFile := fs.NewFile(t, "", fs.WithContent(cfgFileContents))
	defer cfgFile.Remove()

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
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// default to no config for irrelevant tests
			if tt.wd == "" {
				tt.wd = dirEmpty.Path()
			}

			if err = os.Chdir(tt.wd); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(cwd) // nolint: errcheck

			opts := mockOptGetter{
				bools:   tt.bools,
				strings: tt.strings,
			}
			var meta Metadata

			if err = meta.Set(opts); err != nil {
				t.Fatal(err)
			}

			// evaluate symlinks to avoid noise
			if meta.Directory, err = filepath.EvalSymlinks(meta.Directory); err != nil {
				t.Fatal(err)
			}
			if tt.meta.Directory, err = filepath.EvalSymlinks(tt.meta.Directory); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.meta, meta, compareLoggers); diff != "" {
				t.Errorf("metadata differs:\n%s", diff)
			}
		})
	}
}

var compareLoggers = cmp.Comparer(func(a, b *ui.Logger) bool {
	return a.Stderr == b.Stderr && a.Stdout == b.Stdout && a.Verbosity == b.Verbosity
})
