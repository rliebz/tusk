package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestSearchForFile(t *testing.T) {
	tmpdir, cleanup := useTempDir(t)
	defer cleanup()

	emptyDir := mkDir(t, tmpdir, "empty")

	topDir := mkDir(t, tmpdir, "top")
	topConfig := mkConfigFile(t, topDir, "tusk.yml")

	yamlDir := mkDir(t, tmpdir, "yaml")
	yamlConfig := mkConfigFile(t, yamlDir, "tusk.yaml")

	nestedDir := mkDir(t, topDir, "foo", "bar")
	nestedConfig := mkConfigFile(t, nestedDir, "tusk.yml")

	inheritedDir := mkDir(t, topDir, "baz", "empty")

	testcases := []struct {
		wd    string
		path  string
		found bool
	}{
		{
			wd:    emptyDir,
			path:  "",
			found: false,
		},
		{
			wd:    topDir,
			path:  topConfig,
			found: true,
		},
		{
			wd:    yamlDir,
			path:  yamlConfig,
			found: true,
		},
		{
			wd:    nestedDir,
			path:  nestedConfig,
			found: true,
		},
		{
			wd:    inheritedDir,
			path:  topConfig,
			found: true,
		},
	}

	for _, tt := range testcases {
		if err := os.Chdir(tt.wd); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		fullPath, found, err := searchForFile()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}

		if tt.path != fullPath {
			t.Errorf(
				"SearchForFile(): expected path: %s, actual: %s",
				tt.path, fullPath,
			)
		}

		if tt.found != found {
			t.Errorf(
				"SearchForFile(): expected found: %v, actual: %v",
				tt.found, found,
			)
		}
	}
}

func useTempDir(t *testing.T) (string, func()) {
	t.Helper()

	tmpdir, err := ioutil.TempDir("", "tusk-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	tmpdir, err = filepath.EvalSymlinks(tmpdir)
	if err != nil {
		t.Fatalf("failed to follow symlink for temp dir: %v", err)
	}

	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmpdir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(tmpdir); err != nil {
			t.Logf("failed to remove tmpdir: %v", err)
		}
		if err := os.Chdir(oldwd); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}
	}

	return tmpdir, cleanup
}

func mkDir(t *testing.T, elem ...string) string {
	t.Helper()

	fullPath := filepath.Join(elem...)
	if err := os.MkdirAll(fullPath, 0750); err != nil {
		t.Fatalf("failed to make directory: %v", err)
	}

	return fullPath
}

func mkConfigFile(t *testing.T, dir, fileName string) string {
	t.Helper()

	fullPath := filepath.Join(dir, fileName)
	if err := ioutil.WriteFile(fullPath, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	return fullPath
}
