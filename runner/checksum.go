package runner

import (
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/rliebz/tusk/internal/xdg"
)

func (t *Task) isUpToDate(ctx Context) (bool, error) {
	cachePath, err := t.taskInputCachePath(ctx)
	if err != nil {
		return false, err
	}

	outputChecksum, err := t.outputChecksum(ctx)
	if err != nil {
		return false, err
	}

	// TODO: A non-timestamp based implementation
	fmt.Println("cache location:", cachePath)
	fmt.Println("output checksum:", outputChecksum)
	return t.isUpToDateTS(ctx)
}

// taskInputCachePath returns a unique file path based on the inputs of a task.
func (t *Task) taskInputCachePath(ctx Context) (string, error) {
	taskCacheDir, err := t.taskCacheDir(ctx)
	if err != nil {
		return "", err
	}

	h := fnv.New64a()

	for _, glob := range t.Source {
		err := doublestar.GlobWalk(
			os.DirFS(ctx.Dir()),
			glob,
			func(path string, d fs.DirEntry) error {
				return hashFile(h, path, d)
			},
			doublestar.WithFailOnIOErrors(),
			doublestar.WithFilesOnly(),
			doublestar.WithFailOnPatternNotExist(),
		)
		switch {
		// If a source pattern does not exist, that's an error
		case errors.Is(err, doublestar.ErrPatternNotExist):
			return "", fmt.Errorf("no source files found matching pattern: %s", glob)
		case err != nil:
			return "", err
		}
	}

	filename := encodeToString(h)
	return filepath.Join(taskCacheDir, filename), nil
}

// taskCacheDir returns the file path specific to this task.
//
// It incorporates
func (t *Task) taskCacheDir(ctx Context) (string, error) {
	projectCacheDir, err := projectCacheDir(ctx)
	if err != nil {
		return "", err
	}

	h := fnv.New64a()
	if _, err := io.WriteString(h, t.Name); err != nil {
		return "", err
	}

	filename := encodeToString(h)
	return filepath.Join(projectCacheDir, filename), nil
}

func projectCacheDir(ctx Context) (string, error) {
	xdgCacheHome, err := xdg.CacheHome()
	if err != nil {
		return "", err
	}

	h := fnv.New64a()

	if _, err := io.WriteString(h, ctx.CfgPath); err != nil {
		return "", err
	}

	filename := encodeToString(h)
	return filepath.Join(xdgCacheHome, "tusk", filename), nil
}

// outputChecksum returns a checksum for the output of a task.
func (t *Task) outputChecksum(ctx Context) (string, error) {
	h := fnv.New64a()

	for _, glob := range t.Target {
		err := doublestar.GlobWalk(
			os.DirFS(ctx.Dir()),
			glob,
			func(path string, d fs.DirEntry) error {
				return hashFile(h, path, d)
			},
			doublestar.WithFailOnIOErrors(),
			doublestar.WithFilesOnly(),
			doublestar.WithFailOnPatternNotExist(),
		)
		switch {
		// If a target pattern does not exist, we're not up to date.
		case errors.Is(err, doublestar.ErrPatternNotExist):
			return "", nil
		case err != nil:
			return "", err
		}
	}

	filename := base64.RawStdEncoding.EncodeToString(h.Sum(nil))
	return filename, nil
}

func hashFile(hasher io.Writer, path string, d fs.DirEntry) error {
	if _, err := io.WriteString(hasher, path); err != nil {
		return err
	}

	if _, err := io.WriteString(hasher, d.Type().String()); err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close() //nolint:errcheck

	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}

	return nil
}

func encodeToString(h hash.Hash) string {
	return base64.RawStdEncoding.EncodeToString(h.Sum(nil))
}
