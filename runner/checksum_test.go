package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
)

func TestCleanCache(t *testing.T) {
	g := ghost.New(t)

	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	cacheDir := filepath.Join(cacheHome, "tusk")
	err := os.MkdirAll(cacheDir, 0o700)
	g.NoError(err)

	err = CleanCache()
	g.NoError(err)

	_, err = os.Stat(cacheDir)
	g.Should(be.ErrorIs(err, os.ErrNotExist))
}

func TestCleanProjectCache(t *testing.T) {
	g := ghost.New(t)

	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	cacheDir := filepath.Join(cacheHome, "tusk")

	projectDir1, err := projectCacheDir("tusk.yml")
	g.NoError(err)
	err = os.MkdirAll(projectDir1, 0o700)
	g.NoError(err)

	projectDir2, err := projectCacheDir("tusk-2.yml")
	g.NoError(err)
	err = os.MkdirAll(projectDir2, 0o700)
	g.NoError(err)

	entries, err := os.ReadDir(cacheDir)
	g.NoError(err)
	g.Must(be.SliceLen(entries, 2))

	err = CleanProjectCache("tusk.yml")
	g.NoError(err)

	entries, err = os.ReadDir(cacheDir)
	g.NoError(err)
	g.Must(be.SliceLen(entries, 1))
}

func TestCleanProjectCache_no_cfg_path(t *testing.T) {
	g := ghost.New(t)

	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome) // just in case

	err := CleanProjectCache("")
	g.Should(be.ErrorEqual(err, "no config file found"))
}

func TestCleanTaskCache(t *testing.T) {
	g := ghost.New(t)

	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	taskDir, err := taskCacheDir("tusk.yml", "my-task")
	g.NoError(err)
	err = os.MkdirAll(taskDir, 0o700)
	g.NoError(err)

	sameProjectTaskDir, err := taskCacheDir("tusk.yml", "other-task")
	g.NoError(err)
	err = os.MkdirAll(sameProjectTaskDir, 0o700)
	g.NoError(err)

	otherProjectTaskDir, err := taskCacheDir("tusk-2.yml", "my-task")
	g.NoError(err)
	err = os.MkdirAll(otherProjectTaskDir, 0o700)
	g.NoError(err)

	projectDir, err := projectCacheDir("tusk.yml")
	g.NoError(err)

	entries, err := os.ReadDir(projectDir)
	g.NoError(err)
	g.Must(be.SliceLen(entries, 2))

	err = CleanTaskCache("tusk.yml", "my-task")
	g.NoError(err)

	entries, err = os.ReadDir(projectDir)
	g.NoError(err)
	g.Must(be.SliceLen(entries, 1))

	otherProjectDir, err := projectCacheDir("tusk-2.yml")
	g.NoError(err)

	otherEntries, err := os.ReadDir(otherProjectDir)
	g.NoError(err)
	g.Must(be.SliceLen(otherEntries, 1))
}

func TestCleanTaskCache_no_cfg_path(t *testing.T) {
	g := ghost.New(t)

	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome) // just in case

	err := CleanTaskCache("", "foo")
	g.Should(be.ErrorEqual(err, "no config file found"))
}
