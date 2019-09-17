package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rliebz/tusk/ui"
	"gotest.tools/fs"
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
				Verbosity: ui.VerbosityLevelNormal,
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
				Verbosity: ui.VerbosityLevelNormal,
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
				Verbosity: ui.VerbosityLevelNormal,
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
				Verbosity: ui.VerbosityLevelNormal,
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
				Verbosity:         ui.VerbosityLevelNormal,
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
				Verbosity:           ui.VerbosityLevelNormal,
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
				Verbosity: ui.VerbosityLevelNormal,
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
				Verbosity:    ui.VerbosityLevelNormal,
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
				Verbosity: ui.VerbosityLevelSilent,
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
				Verbosity: ui.VerbosityLevelQuiet,
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
				Verbosity: ui.VerbosityLevelVerbose,
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
				Verbosity: ui.VerbosityLevelSilent,
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

			if diff := cmp.Diff(tt.meta, meta); diff != "" {
				t.Errorf("metadata differs:\n%s", diff)
			}
		})
	}
}
