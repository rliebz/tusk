package appcli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/rliebz/tusk/runner"
	"github.com/rliebz/tusk/ui"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/env"
	"gotest.tools/v3/fs"
)

func TestInstallCompletionUnsupported(t *testing.T) {
	err := InstallCompletion(
		&runner.Metadata{
			InstallCompletion: "fake",
		},
	)
	assert.ErrorContains(t, err, `tab completion for "fake" is not supported`)
}

func TestUninstallCompletionUnsupported(t *testing.T) {
	err := UninstallCompletion(
		&runner.Metadata{
			UninstallCompletion: "fake",
		},
	)
	assert.ErrorContains(t, err, `tab completion for "fake" is not supported`)
}

func TestInstallBashCompletion(t *testing.T) {
	homedir := fs.NewDir(t, "home")
	defer homedir.Remove()

	datadir := fs.NewDir(t, "data", fs.WithDir("tusk"))
	defer datadir.Remove()

	defer env.PatchAll(t, map[string]string{
		"HOME":          homedir.Path(),
		"USERPROFILE":   homedir.Path(),
		"XDG_DATA_HOME": datadir.Path(),
	})()

	err := installBashCompletion(ui.Noop())
	assert.NilError(t, err)

	completionFile := filepath.Join(datadir.Path(), "tusk", "tusk-completion.bash")
	contents, err := os.ReadFile(completionFile)
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal(string(contents), rawBashCompletion))

	rcfile := filepath.Join(homedir.Path(), ".bashrc")
	rcContents, err := os.ReadFile(rcfile)
	assert.NilError(t, err)

	command := fmt.Sprintf("source %q", filepath.ToSlash(completionFile))
	assert.Check(t, cmp.Contains(string(rcContents), command))
}

func TestGetBashRCFile(t *testing.T) {
	tests := []struct {
		name   string
		ops    []fs.PathOp
		expect string
	}{
		{
			"no-files",
			[]fs.PathOp{},
			".bashrc",
		},
		{
			"all-files",
			[]fs.PathOp{
				fs.WithFile(".bashrc", ""),
				fs.WithFile(".bash_profile", ""),
				fs.WithFile(".profile", ""),
			},
			".bashrc",
		},
		{
			"bash-profile-only",
			[]fs.PathOp{
				fs.WithFile(".bash_profile", ""),
			},
			".bash_profile",
		},
		{
			"profile-only",
			[]fs.PathOp{
				fs.WithFile(".profile", ""),
			},
			".profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			homedir := fs.NewDir(t, "home", tt.ops...)
			defer homedir.Remove()

			defer env.PatchAll(t, map[string]string{
				"HOME":        homedir.Path(),
				"USERPROFILE": homedir.Path(),
			})()

			rcfile, err := getBashRCFile()
			assert.NilError(t, err)

			want := filepath.Join(homedir.Path(), tt.expect)
			assert.Check(t, cmp.Equal(want, rcfile))
		})
	}
}

func TestAppendIfAbsent_trailing_newlines(t *testing.T) {
	existing := "# First Line\n\n"
	f := fs.NewFile(t, "bashrc", fs.WithContent(existing))
	defer f.Remove()

	text := "# Second Line"
	err := appendIfAbsent(f.Path(), text)
	assert.NilError(t, err)

	want := existing + text + "\n"
	got, err := os.ReadFile(f.Path())
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal(want, string(got)))
}

func TestAppendIfAbsent_no_trailing_newline(t *testing.T) {
	existing := "# First Line"
	f := fs.NewFile(t, "bashrc", fs.WithContent(existing))
	defer f.Remove()

	text := "# Second Line"
	err := appendIfAbsent(f.Path(), text)
	assert.NilError(t, err)

	want := existing + "\n" + text + "\n"
	got, err := os.ReadFile(f.Path())
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal(want, string(got)))
}

func TestAppendIfAbsent_exists(t *testing.T) {
	text := "# Existing Line"
	f := fs.NewFile(t, "bashrc", fs.WithContent(text))
	defer f.Remove()

	err := appendIfAbsent(f.Path(), text)
	assert.NilError(t, err)

	got, err := os.ReadFile(f.Path())
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal(text, string(got)))
}

func TestAppendIfAbsent_no_file(t *testing.T) {
	f := fs.NewFile(t, "bashrc")
	defer f.Remove() // Will be recreated

	f.Remove()

	text := "# Target Line"
	err := appendIfAbsent(f.Path(), text)
	assert.NilError(t, err)

	got, err := os.ReadFile(f.Path())
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal(text+"\n", string(got)))
}

func TestUninstallBashCompletion(t *testing.T) {
	datadir := fs.NewDir(
		t,
		"data",
		fs.WithDir("tusk",
			fs.WithFile("tusk-completion.bash", rawBashCompletion),
		),
	)
	defer datadir.Remove()

	rcfile := filepath.Join(datadir.Path(), "tusk", "tusk-completion.bash")

	contents := fmt.Sprintf("# Preamble\nsource %q", filepath.ToSlash(rcfile))
	homedir := fs.NewDir(t, "home", fs.WithFile(".bashrc", contents))
	defer homedir.Remove()

	defer env.PatchAll(t, map[string]string{
		"HOME":          homedir.Path(),
		"USERPROFILE":   homedir.Path(),
		"XDG_DATA_HOME": datadir.Path(),
	})()

	err := uninstallBashCompletion()
	assert.NilError(t, err)

	_, err = os.Stat(rcfile)
	assert.Check(t, os.IsNotExist(err))

	got, err := os.ReadFile(filepath.Join(homedir.Path(), ".bashrc"))
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal("# Preamble\n", string(got)))
}

func TestRemoveLineInFile(t *testing.T) {
	content := `# First
match

# Second

match`
	want := `# First

# Second
`

	file := fs.NewFile(t, "file", fs.WithContent(content))
	defer file.Remove()

	err := removeLineInFile(file.Path(), regexp.MustCompile("match"))
	assert.NilError(t, err)

	got, err := os.ReadFile(file.Path())
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal(want, string(got)))
}

func TestInstallFishCompletion(t *testing.T) {
	cfgdir := fs.NewDir(t, "data")
	defer cfgdir.Remove()

	defer env.PatchAll(t, map[string]string{
		"XDG_CONFIG_HOME": cfgdir.Path(),
	})()

	err := installFishCompletion(ui.Noop())
	assert.NilError(t, err)

	completionFile := filepath.Join(cfgdir.Path(), "fish", "completions", "tusk.fish")
	contents, err := os.ReadFile(completionFile)
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal(string(contents), rawFishCompletion))
}

func TestUninstallFishCompletion(t *testing.T) {
	cfgdir := fs.NewDir(
		t,
		"data",
		fs.WithDir(
			"fish",
			fs.WithDir(
				"completions",
				fs.WithFile("tusk.fish", rawFishCompletion),
			),
		),
	)
	defer cfgdir.Remove()

	completionFile := filepath.Join(cfgdir.Path(), "fish", "completions", "tusk.fish")
	_, err := os.Stat(completionFile)
	assert.NilError(t, err)

	defer env.PatchAll(t, map[string]string{
		"XDG_CONFIG_HOME": cfgdir.Path(),
	})()

	err = uninstallFishCompletion()
	assert.NilError(t, err)

	_, err = os.Stat(completionFile)
	assert.Check(t, os.IsNotExist(err))
}

func TestGetDataDir_xdg(t *testing.T) {
	xdgDataHome := "/foo/bar/baz"
	defer env.Patch(t, "XDG_DATA_HOME", xdgDataHome)()

	want := filepath.Join(xdgDataHome, "tusk")

	got, err := getDataDir()
	assert.NilError(t, err)

	assert.Equal(t, want, got)
}

func TestGetDataDir_default(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.NilError(t, err)

	want := filepath.Join(home, ".local", "share", "tusk")

	got, err := getDataDir()
	assert.NilError(t, err)

	assert.Equal(t, want, got)
}

func TestGetFishCompletionsDir_xdg(t *testing.T) {
	cfgHome := "/foo/bar/baz"
	defer env.Patch(t, "XDG_CONFIG_HOME", cfgHome)()

	want := filepath.Join(cfgHome, "fish", "completions")

	got, err := getFishCompletionsDir()
	assert.NilError(t, err)

	assert.Equal(t, want, got)
}

func TestGetFishCompletionsDir_default(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.NilError(t, err)

	want := filepath.Join(home, ".config", "fish", "completions")

	got, err := getFishCompletionsDir()
	assert.NilError(t, err)

	assert.Equal(t, want, got)
}

func TestInstallZshCompletion(t *testing.T) {
	dir := fs.NewDir(t, "project-dir")
	defer dir.Remove()

	err := installZshCompletion(ui.Noop(), dir.Path())
	assert.NilError(t, err)

	contents, err := os.ReadFile(filepath.Join(dir.Path(), "_tusk"))
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal(string(contents), rawZshCompletion))
}

func TestUninstallZshCompletion(t *testing.T) {
	dir := fs.NewDir(t, "project-dir", fs.WithFile("_tusk", rawZshCompletion))
	defer dir.Remove()

	err := uninstallZshCompletion(dir.Path())
	assert.NilError(t, err)

	_, err = os.Stat(filepath.Join(dir.Path(), "_tusk"))
	assert.Assert(t, os.IsNotExist(err))
}

func TestUninstallZshCompletion_empty(t *testing.T) {
	dir := fs.NewDir(t, "project-dir")
	defer dir.Remove()

	err := uninstallZshCompletion(dir.Path())
	assert.NilError(t, err)

	_, err = os.Stat(filepath.Join(dir.Path(), "_tusk"))
	assert.Assert(t, os.IsNotExist(err))
}
