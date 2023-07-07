package appcli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	"gotest.tools/v3/fs"

	"github.com/rliebz/tusk/runner"
	"github.com/rliebz/tusk/ui"
)

func TestInstallCompletionUnsupported(t *testing.T) {
	g := ghost.New(t)

	err := InstallCompletion(
		&runner.Metadata{
			InstallCompletion: "fake",
		},
	)
	g.Should(be.ErrorContaining(`tab completion for "fake" is not supported`, err))
}

func TestUninstallCompletionUnsupported(t *testing.T) {
	g := ghost.New(t)

	err := UninstallCompletion(
		&runner.Metadata{
			UninstallCompletion: "fake",
		},
	)
	g.Should(be.ErrorContaining(`tab completion for "fake" is not supported`, err))
}

func TestInstallBashCompletion(t *testing.T) {
	g := ghost.New(t)

	homedir := fs.NewDir(t, "home")
	datadir := fs.NewDir(t, "data", fs.WithDir("tusk"))

	t.Setenv("HOME", homedir.Path())
	t.Setenv("USERPROFILE", homedir.Path())
	t.Setenv("XDG_DATA_HOME", datadir.Path())

	err := installBashCompletion(ui.Noop())
	g.NoError(err)

	completionFile := filepath.Join(datadir.Path(), "tusk", "tusk-completion.bash")
	contents, err := os.ReadFile(completionFile)
	g.NoError(err)

	g.Should(be.Equal(string(contents), rawBashCompletion))

	rcfile := filepath.Join(homedir.Path(), ".bashrc")
	rcContents, err := os.ReadFile(rcfile)
	g.NoError(err)

	wantCommand := fmt.Sprintf("source %q", filepath.ToSlash(completionFile))
	g.Should(be.ContainingString(wantCommand, string(rcContents)))
}

func TestGetBashRCFile(t *testing.T) {
	tests := []struct {
		name string
		ops  []fs.PathOp
		want string
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
			g := ghost.New(t)

			homedir := fs.NewDir(t, "home", tt.ops...)

			t.Setenv("HOME", homedir.Path())
			t.Setenv("USERPROFILE", homedir.Path())

			rcfile, err := getBashRCFile()
			g.NoError(err)

			want := filepath.Join(homedir.Path(), tt.want)
			g.Should(be.Equal(want, rcfile))
		})
	}
}

func TestAppendIfAbsent(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		append   string
		want     string
	}{
		{
			name:     "trailing newlines",
			existing: "# First Line\n\n",
			append:   "# Second Line",
			want:     "# First Line\n\n# Second Line\n",
		},
		{
			name:     "no trailing newlines",
			existing: "# First Line",
			append:   "# Second Line",
			want:     "# First Line\n# Second Line\n",
		},
		{
			name:     "empty file",
			existing: "",
			append:   "# New Line",
			want:     "# New Line\n",
		},
		{
			name:     "exists",
			existing: "# Existing Line",
			append:   "# Existing Line",
			want:     "# Existing Line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			f := fs.NewFile(t, "bashrc", fs.WithContent(tt.existing))

			err := appendIfAbsent(f.Path(), tt.append)
			g.NoError(err)

			got, err := os.ReadFile(f.Path())
			g.NoError(err)

			g.Should(be.Equal(tt.want, string(got)))
		})
	}

	t.Run("no file", func(t *testing.T) {
		g := ghost.New(t)

		f := fs.NewFile(t, "bashrc")

		text := "# Target Line"
		err := appendIfAbsent(f.Path(), text)
		g.NoError(err)

		got, err := os.ReadFile(f.Path())
		g.NoError(err)

		g.Should(be.Equal(text+"\n", string(got)))
	})
}

func TestUninstallBashCompletion(t *testing.T) {
	g := ghost.New(t)

	datadir := fs.NewDir(
		t,
		"data",
		fs.WithDir("tusk",
			fs.WithFile("tusk-completion.bash", rawBashCompletion),
		),
	)

	rcfile := filepath.Join(datadir.Path(), "tusk", "tusk-completion.bash")

	contents := fmt.Sprintf("# Preamble\nsource %q", filepath.ToSlash(rcfile))
	homedir := fs.NewDir(t, "home", fs.WithFile(".bashrc", contents))

	t.Setenv("HOME", homedir.Path())
	t.Setenv("USERPROFILE", homedir.Path())
	t.Setenv("XDG_DATA_HOME", datadir.Path())

	err := uninstallBashCompletion()
	g.NoError(err)

	_, err = os.Stat(rcfile)
	g.Should(be.True(os.IsNotExist(err)))

	got, err := os.ReadFile(filepath.Join(homedir.Path(), ".bashrc"))
	g.NoError(err)

	g.Should(be.Equal("# Preamble\n", string(got)))
}

func TestRemoveLineInFile(t *testing.T) {
	g := ghost.New(t)

	content := `# First
match

# Second

match`
	want := `# First

# Second
`

	file := fs.NewFile(t, "file", fs.WithContent(content))

	err := removeLineInFile(file.Path(), regexp.MustCompile("match"))
	g.NoError(err)

	got, err := os.ReadFile(file.Path())
	g.NoError(err)

	g.Should(be.Equal(want, string(got)))
}

func TestInstallFishCompletion(t *testing.T) {
	g := ghost.New(t)

	cfgdir := fs.NewDir(t, "data")

	t.Setenv("XDG_CONFIG_HOME", cfgdir.Path())

	err := installFishCompletion(ui.Noop())
	g.NoError(err)

	completionFile := filepath.Join(cfgdir.Path(), "fish", "completions", "tusk.fish")
	got, err := os.ReadFile(completionFile)
	g.NoError(err)

	g.Should(be.Equal(rawFishCompletion, string(got)))
}

func TestUninstallFishCompletion(t *testing.T) {
	g := ghost.New(t)

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

	completionFile := filepath.Join(cfgdir.Path(), "fish", "completions", "tusk.fish")
	_, err := os.Stat(completionFile)
	g.NoError(err)

	t.Setenv("XDG_CONFIG_HOME", cfgdir.Path())

	err = uninstallFishCompletion()
	g.NoError(err)

	_, err = os.Stat(completionFile)
	g.Should(be.True(os.IsNotExist(err)))
}

func TestGetDataDir_xdg(t *testing.T) {
	g := ghost.New(t)

	xdgDataHome := "/foo/bar/baz"
	t.Setenv("XDG_DATA_HOME", xdgDataHome)

	want := filepath.Join(xdgDataHome, "tusk")

	got, err := getDataDir()
	g.NoError(err)

	g.Should(be.Equal(want, got))
}

func TestGetDataDir_default(t *testing.T) {
	g := ghost.New(t)

	home, err := os.UserHomeDir()
	g.NoError(err)

	want := filepath.Join(home, ".local", "share", "tusk")

	got, err := getDataDir()
	g.NoError(err)

	g.Should(be.Equal(want, got))
}

func TestGetFishCompletionsDir_xdg(t *testing.T) {
	g := ghost.New(t)

	cfgHome := "/foo/bar/baz"
	t.Setenv("XDG_CONFIG_HOME", cfgHome)

	want := filepath.Join(cfgHome, "fish", "completions")

	got, err := getFishCompletionsDir()
	g.NoError(err)

	g.Should(be.Equal(want, got))
}

func TestGetFishCompletionsDir_default(t *testing.T) {
	g := ghost.New(t)

	home, err := os.UserHomeDir()
	g.NoError(err)

	want := filepath.Join(home, ".config", "fish", "completions")

	got, err := getFishCompletionsDir()
	g.NoError(err)

	g.Should(be.Equal(want, got))
}

func TestInstallZshCompletion(t *testing.T) {
	g := ghost.New(t)

	dir := fs.NewDir(t, "project-dir")

	err := installZshCompletion(ui.Noop(), dir.Path())
	g.NoError(err)

	contents, err := os.ReadFile(filepath.Join(dir.Path(), "_tusk"))
	g.NoError(err)

	g.Should(be.Equal(rawZshCompletion, string(contents)))
}

func TestUninstallZshCompletion(t *testing.T) {
	g := ghost.New(t)

	dir := fs.NewDir(t, "project-dir", fs.WithFile("_tusk", rawZshCompletion))

	err := uninstallZshCompletion(dir.Path())
	g.NoError(err)

	_, err = os.Stat(filepath.Join(dir.Path(), "_tusk"))
	g.Should(be.True(os.IsNotExist(err)))
}

func TestUninstallZshCompletion_empty(t *testing.T) {
	g := ghost.New(t)

	dir := fs.NewDir(t, "project-dir")

	err := uninstallZshCompletion(dir.Path())
	g.NoError(err)

	_, err = os.Stat(filepath.Join(dir.Path(), "_tusk"))
	g.Should(be.True(os.IsNotExist(err)))
}
