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

	"github.com/rliebz/tusk/ui"
)

func TestInstallCompletionUnsupported(t *testing.T) {
	g := ghost.New(t)

	err := InstallCompletion(
		&Metadata{
			InstallCompletion: "fake",
		},
	)
	g.Should(be.ErrorEqual(err,
		`completion target "fake" must be one of [bash, fish, zsh]`,
	))
}

func TestUninstallCompletionUnsupported(t *testing.T) {
	g := ghost.New(t)

	err := UninstallCompletion(
		&Metadata{
			UninstallCompletion: "fake",
		},
	)
	g.Should(be.ErrorEqual(err,
		`completion target "fake" must be one of [bash, fish, zsh]`,
	))
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

	g.Should(be.Equal(rawBashCompletion, string(contents)))

	rcfile := filepath.Join(homedir.Path(), ".bashrc")
	rcContents, err := os.ReadFile(rcfile)
	g.NoError(err)

	wantCommand := fmt.Sprintf("source %q", filepath.ToSlash(completionFile))
	g.Should(be.StringContaining(string(rcContents), wantCommand))
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
			g.Should(be.Equal(rcfile, want))
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

			g.Should(be.Equal(string(got), tt.want))
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

		g.Should(be.Equal(string(got), text+"\n"))
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
	g.Should(be.ErrorIs(err, os.ErrNotExist))

	got, err := os.ReadFile(filepath.Join(homedir.Path(), ".bashrc"))
	g.NoError(err)

	g.Should(be.Equal(string(got), "# Preamble\n"))
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

	g.Should(be.Equal(string(got), want))
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

	g.Should(be.Equal(string(got), rawFishCompletion))
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
	g.Should(be.ErrorIs(err, os.ErrNotExist))
}

func TestGetDataDir_xdg(t *testing.T) {
	g := ghost.New(t)

	xdgDataHome := "/foo/bar/baz"
	t.Setenv("XDG_DATA_HOME", xdgDataHome)

	want := filepath.Join(xdgDataHome, "tusk")

	got, err := getDataDir()
	g.NoError(err)

	g.Should(be.Equal(got, want))
}

func TestGetDataDir_default(t *testing.T) {
	g := ghost.New(t)

	home, err := os.UserHomeDir()
	g.NoError(err)

	want := filepath.Join(home, ".local", "share", "tusk")

	got, err := getDataDir()
	g.NoError(err)

	g.Should(be.Equal(got, want))
}

func TestGetFishCompletionsDir_xdg(t *testing.T) {
	g := ghost.New(t)

	cfgHome := "/foo/bar/baz"
	t.Setenv("XDG_CONFIG_HOME", cfgHome)

	want := filepath.Join(cfgHome, "fish", "completions")

	got, err := getFishCompletionsDir()
	g.NoError(err)

	g.Should(be.Equal(got, want))
}

func TestGetFishCompletionsDir_default(t *testing.T) {
	g := ghost.New(t)

	home, err := os.UserHomeDir()
	g.NoError(err)

	want := filepath.Join(home, ".config", "fish", "completions")

	got, err := getFishCompletionsDir()
	g.NoError(err)

	g.Should(be.Equal(got, want))
}

func TestInstallZshCompletion(t *testing.T) {
	g := ghost.New(t)

	dir := fs.NewDir(t, "project-dir")

	err := installZshCompletion(ui.Noop(), dir.Path())
	g.NoError(err)

	contents, err := os.ReadFile(filepath.Join(dir.Path(), "_tusk"))
	g.NoError(err)

	g.Should(be.Equal(string(contents), rawZshCompletion))
}

func TestUninstallZshCompletion(t *testing.T) {
	g := ghost.New(t)

	dir := fs.NewDir(t, "project-dir", fs.WithFile("_tusk", rawZshCompletion))

	err := uninstallZshCompletion(dir.Path())
	g.NoError(err)

	_, err = os.Stat(filepath.Join(dir.Path(), "_tusk"))
	g.Should(be.ErrorIs(err, os.ErrNotExist))
}

func TestUninstallZshCompletion_empty(t *testing.T) {
	g := ghost.New(t)

	dir := fs.NewDir(t, "project-dir")

	err := uninstallZshCompletion(dir.Path())
	g.NoError(err)

	_, err = os.Stat(filepath.Join(dir.Path(), "_tusk"))
	g.Should(be.ErrorIs(err, os.ErrNotExist))
}
