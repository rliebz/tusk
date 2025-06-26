package appcli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"

	"github.com/rliebz/tusk/internal/xtesting"
)

func Test_searchForFile(t *testing.T) {
	tmpdir := xtesting.UseTempDir(t)

	emptyDir := mkDir(t, tmpdir, "empty")

	topDir := mkDir(t, tmpdir, "top")
	topConfig := mkConfigFile(t, topDir, "tusk.yml")

	yamlDir := mkDir(t, tmpdir, "yaml")
	yamlConfig := mkConfigFile(t, yamlDir, "tusk.yaml")

	nestedDir := mkDir(t, topDir, "foo", "bar")
	nestedConfig := mkConfigFile(t, nestedDir, "tusk.yml")

	inheritedDir := mkDir(t, topDir, "baz", "empty")

	tests := []struct {
		wd       string
		wantPath string
	}{
		{
			wd:       emptyDir,
			wantPath: "",
		},
		{
			wd:       topDir,
			wantPath: topConfig,
		},
		{
			wd:       yamlDir,
			wantPath: yamlConfig,
		},
		{
			wd:       nestedDir,
			wantPath: nestedConfig,
		},
		{
			wd:       inheritedDir,
			wantPath: topConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.wd+"_"+tt.wantPath, func(t *testing.T) {
			g := ghost.New(t)

			t.Chdir(tt.wd)

			fullPath, err := searchForFile()
			g.NoError(err)

			g.Should(be.Equal(fullPath, tt.wantPath))
		})
	}
}

func mkDir(t *testing.T, elem ...string) string {
	t.Helper()

	g := ghost.New(t)

	fullPath := filepath.Join(elem...)
	err := os.MkdirAll(fullPath, 0o700)
	g.NoError(err)

	return fullPath
}

func mkConfigFile(t *testing.T, dir, fileName string) string {
	t.Helper()

	g := ghost.New(t)

	fullPath := filepath.Join(dir, fileName)
	err := os.WriteFile(fullPath, []byte{}, 0o600)
	g.NoError(err)

	return fullPath
}
