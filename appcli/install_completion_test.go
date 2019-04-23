package appcli

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	"gotest.tools/env"
	"gotest.tools/fs"
)

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

func TestAppendIfAbsent(t *testing.T) {
	existing := "# First Line\n"
	f := fs.NewFile(t, "bashrc", fs.WithContent(existing))
	defer f.Remove()

	text := "# Second Line"
	err := appendIfAbsent(f.Path(), text)
	assert.NilError(t, err)

	want := existing + text + "\n"
	got, err := ioutil.ReadFile(f.Path())
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal(want, string(got)))
}

func TestAppendIfAbsent_exists(t *testing.T) {
	text := "# Existing Line"
	f := fs.NewFile(t, "bashrc", fs.WithContent(text))
	defer f.Remove()

	err := appendIfAbsent(f.Path(), text)
	assert.NilError(t, err)

	got, err := ioutil.ReadFile(f.Path())
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

	got, err := ioutil.ReadFile(f.Path())
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal(text+"\n", string(got)))
}

func TestInstallZshCompletion(t *testing.T) {
	dir := fs.NewDir(t, "project-dir")
	defer dir.Remove()

	err := installZshCompletion(dir.Path())
	assert.NilError(t, err)

	contents, err := ioutil.ReadFile(filepath.Join(dir.Path(), "_tusk"))
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
