package appcli

import (
	"io/ioutil"
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
	if err != nil {
		t.Fatal(err)
	}

	contents, err := ioutil.ReadFile(filepath.Join(dir.Path(), "_tusk"))
	if err != nil {
		t.Fatal(err)
	}

	assert.Check(t, cmp.Equal(string(contents), rawZshCompletion))
}
