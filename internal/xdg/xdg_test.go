package xdg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
)

func TestCacheHome_env(t *testing.T) {
	g := ghost.New(t)

	xdgCacheHome := "/foo/bar/baz"
	t.Setenv("XDG_CACHE_HOME", xdgCacheHome)

	got, err := CacheHome()
	g.NoError(err)

	g.Should(be.Equal(got, xdgCacheHome))
}

func TestCacheHome_default(t *testing.T) {
	g := ghost.New(t)

	home, err := os.UserHomeDir()
	g.NoError(err)

	want := filepath.Join(home, ".cache")

	got, err := CacheHome()
	g.NoError(err)

	g.Should(be.Equal(got, want))
}

func TestConfigHome_env(t *testing.T) {
	g := ghost.New(t)

	xdgConfigHome := "/foo/bar/baz"
	t.Setenv("XDG_CONFIG_HOME", xdgConfigHome)

	got, err := ConfigHome()
	g.NoError(err)

	g.Should(be.Equal(got, xdgConfigHome))
}

func TestConfigHome_default(t *testing.T) {
	g := ghost.New(t)

	home, err := os.UserHomeDir()
	g.NoError(err)

	want := filepath.Join(home, ".config")

	got, err := ConfigHome()
	g.NoError(err)

	g.Should(be.Equal(got, want))
}

func TestDataHome_env(t *testing.T) {
	g := ghost.New(t)

	xdgDataHome := "/foo/bar/baz"
	t.Setenv("XDG_DATA_HOME", xdgDataHome)

	got, err := DataHome()
	g.NoError(err)

	g.Should(be.Equal(got, xdgDataHome))
}

func TestDataHome_default(t *testing.T) {
	g := ghost.New(t)

	home, err := os.UserHomeDir()
	g.NoError(err)

	want := filepath.Join(home, ".local", "share")

	got, err := DataHome()
	g.NoError(err)

	g.Should(be.Equal(got, want))
}
