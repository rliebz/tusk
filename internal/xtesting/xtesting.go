package xtesting

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
)

// StashEnv clears the environment for the duration of the testing.
func StashEnv(t testing.TB) {
	t.Helper()

	environ := os.Environ()

	t.Cleanup(func() {
		for _, val := range environ {
			parts := strings.Split(val, "=")
			err := os.Setenv(parts[0], parts[1]) //nolint:usetesting
			if err != nil {
				t.Errorf("failed to clean up environment: %s", err)
			}
		}
	})

	os.Clearenv()
}

// UseTempDir creates a temporary directory and switches to it.
func UseTempDir(t *testing.T) string {
	t.Helper()
	g := ghost.New(t)

	// MacOS gets fancy with symlinks, so this gets us the real working path.
	tmpdir, err := filepath.EvalSymlinks(t.TempDir())
	g.NoError(err)

	oldwd, err := os.Getwd()
	g.NoError(err)

	err = os.Chdir(tmpdir)
	g.NoError(err)

	t.Cleanup(func() {
		err := os.Chdir(oldwd)
		g.Should(be.Nil(err))
	})

	return tmpdir
}
