package xtesting

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rliebz/ghost"
)

// StashEnv clears the environment for the duration of the testing.
func StashEnv(t testing.TB) {
	environ := os.Environ()

	t.Cleanup(func() {
		for _, val := range environ {
			parts := strings.Split(val, "=")
			os.Setenv(parts[0], parts[1]) //nolint:errcheck,usetesting
		}
	})

	os.Clearenv()
}

// UseTempDir creates a temporary directory and switches to it.
func UseTempDir(t *testing.T) string {
	g := ghost.New(t)

	// MacOS gets fancy with symlinks, so this gets us the real working path.
	tmpdir, err := filepath.EvalSymlinks(t.TempDir())
	g.NoError(err)

	t.Chdir(tmpdir)

	return tmpdir
}
