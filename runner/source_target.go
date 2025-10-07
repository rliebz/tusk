package runner

import (
	"cmp"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/bmatcuk/doublestar/v4"
	"golang.org/x/sync/errgroup"

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
	if cfgPath == "" {
		return errors.New("no config file found")
	}

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
	if cfgPath == "" {
		return errors.New("no config file found")
	}

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
func (t *Task) taskInputCachePath(c Context) (string, error) {
	taskCacheDir, err := taskCacheDir(c.CfgPath, t.Name)
	if err != nil {
		return "", err
	}

	filename, err := dirChecksum("source", os.DirFS(c.Dir()), t.Source)
	if err != nil {
		return "", err
	}

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
func (t *Task) outputChecksum(c Context) (string, error) {
	filename, err := dirChecksum("target", os.DirFS(c.Dir()), t.Target)
	var pnfe *patternNotFoundError
	switch {
	case errors.As(err, &pnfe):
		return "", nil
	case err != nil:
		return "", err
	}

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

type patternNotFoundError struct {
	kind    string // one of: source, target
	pattern string
}

func (e *patternNotFoundError) Error() string {
	return fmt.Sprintf("no %s files found matching pattern: %s", e.kind, e.pattern)
}

type entry struct {
	path string
	d    fs.DirEntry
}

type result struct {
	path string
	sum  []byte
}

func dirChecksum(kind string, dir fs.FS, patterns []string) (string, error) {
	g, ctx := errgroup.WithContext(context.Background())
	numWorkers := runtime.GOMAXPROCS(0)

	entries := make(chan entry, numWorkers*2)
	g.Go(func() error {
		defer close(entries)
		return walkEntries(ctx, entries, kind, dir, patterns)
	})

	results := make(chan result, numWorkers*2)
	for range numWorkers {
		g.Go(func() error {
			return hashEntries(ctx, results, entries)
		})
	}
	go func() {
		g.Wait() //nolint:errcheck
		close(results)
	}()

	var resultList []result //nolint:prealloc
	for result := range results {
		resultList = append(resultList, result)
	}

	if err := g.Wait(); err != nil {
		return "", err
	}

	slices.SortFunc(resultList, func(a, b result) int {
		return cmp.Compare(a.path, b.path)
	})

	h := fnv.New64a()
	for _, result := range resultList {
		h.Write(result.sum)
	}

	return encodeToString(h), nil
}

// walkEntries iterates over a set of files and writes them to entries.
func walkEntries(
	ctx context.Context,
	entries chan<- entry,
	kind string,
	dir fs.FS,
	patterns []string,
) error {
	for _, glob := range patterns {
		count := 0
		err := doublestar.GlobWalk(
			dir,
			glob,
			func(path string, d fs.DirEntry) error {
				count++
				select {
				case entries <- entry{path, d}:
				case <-ctx.Done():
					return ctx.Err()
				}
				return nil
			},
			doublestar.WithFailOnIOErrors(),
			doublestar.WithFilesOnly(),
			doublestar.WithFailOnPatternNotExist(),
		)
		switch {
		case errors.Is(err, doublestar.ErrPatternNotExist) || (err == nil && count == 0):
			return &patternNotFoundError{kind: kind, pattern: glob}
		case err != nil:
			return err
		}
	}

	return nil
}

// hashEntries iterates over entries and hashes the files into results.
func hashEntries(
	ctx context.Context,
	results chan<- result,
	entries <-chan entry,
) error {
	buf := make([]byte, 1024*1024)
	for entry := range entries {
		sum, err := hashFile(entry.path, entry.d, buf)
		if err != nil {
			return err
		}
		select {
		case results <- result{entry.path, sum}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func hashFile(path string, d fs.DirEntry, buf []byte) ([]byte, error) {
	h := fnv.New64a()
	if _, err := io.WriteString(h, path); err != nil {
		return nil, err
	}

	if _, err := io.WriteString(h, d.Type().String()); err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close() //nolint:errcheck

	if _, err := io.CopyBuffer(h, file, buf); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func encodeToString(h hash.Hash) string {
	return base64.RawStdEncoding.EncodeToString(h.Sum(nil))
}
