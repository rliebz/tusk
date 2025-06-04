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

// CleanCache deletes all cached files.
func CleanCache() error {
	cacheDir, err := tuskCacheDir()
	if err != nil {
		return err
	}

	err = os.RemoveAll(cacheDir)
	if err != nil {
		return fmt.Errorf("cleaning cache dir: %w", err)
	}

	return nil
}

// CleanProjectCache deletes cached files related to the current config file.
func CleanProjectCache(cfgPath string) error {
	cacheDir, err := projectCacheDir(cfgPath)
	if err != nil {
		return err
	}

	err = os.RemoveAll(cacheDir)
	if err != nil {
		return fmt.Errorf("cleaning cache dir: %w", err)
	}

	return nil
}

// CleanTaskCache deletes cached files related to the given task.
func CleanTaskCache(cfgPath string, task string) error {
	cacheDir, err := taskCacheDir(cfgPath, task)
	if err != nil {
		return err
	}

	err = os.RemoveAll(cacheDir)
	if err != nil {
		return fmt.Errorf("cleaning cache dir: %w", err)
	}

	return nil
}

func (t *Task) isUpToDate(ctx Context, cachePath string) (bool, error) {
	if !t.isCacheable() {
		return false, nil
	}

	cachedChecksumBytes, err := os.ReadFile(cachePath)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if len(cachedChecksumBytes) == 0 {
		return false, nil
	}

	outputChecksum, err := t.outputChecksum(ctx)
	if err != nil {
		return false, err
	}

	return outputChecksum == string(cachedChecksumBytes), nil
}

// taskInputCachePath returns a unique file path based on the inputs of a task.
func (t *Task) taskInputCachePath(ctx Context) (string, error) {
	taskCacheDir, err := taskCacheDir(ctx.CfgPath, t.Name)
	if err != nil {
		return "", err
	}

	h := fnv.New64a()

	for _, glob := range t.Source {
		count := 0
		err := doublestar.GlobWalk(
			os.DirFS(ctx.Dir()),
			glob,
			func(path string, d fs.DirEntry) error {
				count++
				return hashFile(h, path, d)
			},
			doublestar.WithFailOnIOErrors(),
			doublestar.WithFilesOnly(),
			doublestar.WithFailOnPatternNotExist(),
		)
		switch {
		// If a source pattern does not exist, that's an error
		case errors.Is(err, doublestar.ErrPatternNotExist) || (err == nil && count == 0):
			return "", fmt.Errorf("no source files found matching pattern: %s", glob)
		case err != nil:
			return "", err
		}
	}

	filename := encodeToString(h)
	return filepath.Join(taskCacheDir, filename), nil
}

// taskCacheDir returns the file path specific to this task.
func taskCacheDir(cfgPath string, taskName string) (string, error) {
	projectCacheDir, err := projectCacheDir(cfgPath)
	if err != nil {
		return "", err
	}

	h := fnv.New64a()
	if _, err := io.WriteString(h, taskName); err != nil {
		return "", err
	}

	filename := encodeToString(h)
	return filepath.Join(projectCacheDir, filename), nil
}

func projectCacheDir(cfgPath string) (string, error) {
	cfgPath, err := filepath.Abs(cfgPath)
	if err != nil {
		return "", err
	}

	cacheDir, err := tuskCacheDir()
	if err != nil {
		return "", err
	}

	h := fnv.New64a()

	if _, err := io.WriteString(h, cfgPath); err != nil {
		return "", err
	}

	filename := encodeToString(h)
	return filepath.Join(cacheDir, filename), nil
}

func tuskCacheDir() (string, error) {
	xdgCacheHome, err := xdg.CacheHome()
	if err != nil {
		return "", err
	}

	return filepath.Join(xdgCacheHome, "tusk"), nil
}

// outputChecksum returns a checksum for the output of a task.
func (t *Task) outputChecksum(ctx Context) (string, error) {
	h := fnv.New64a()

	for _, glob := range t.Target {
		count := 0
		err := doublestar.GlobWalk(
			os.DirFS(ctx.Dir()),
			glob,
			func(path string, d fs.DirEntry) error {
				count++
				return hashFile(h, path, d)
			},
			doublestar.WithFailOnIOErrors(),
			doublestar.WithFilesOnly(),
			doublestar.WithFailOnPatternNotExist(),
		)
		switch {
		// If a target pattern does not exist, we're not up to date.
		case errors.Is(err, doublestar.ErrPatternNotExist) || (err == nil && count == 0):
			return "", nil
		case err != nil:
			return "", err
		}
	}

	filename := base64.RawStdEncoding.EncodeToString(h.Sum(nil))
	return filename, nil
}

func (t *Task) cache(ctx Context, cachePath string) error {
	if !t.isCacheable() {
		return nil
	}

	outputChecksum, err := t.outputChecksum(ctx)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o700); err != nil {
		return err
	}

	if err := os.WriteFile(cachePath, []byte(outputChecksum), 0o600); err != nil {
		return err
	}

	return nil
}

func (t *Task) isCacheable() bool {
	return len(t.Source) != 0 && len(t.Target) != 0
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
