package appcli

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	"gotest.tools/fs"
)

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
