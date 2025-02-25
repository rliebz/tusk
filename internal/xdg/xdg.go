package xdg

import (
	"os"
	"path/filepath"
)

// CacheHome returns the XDG user cache home.
func CacheHome() (string, error) {
	if xdgHome := os.Getenv("XDG_CACHE_HOME"); xdgHome != "" {
		return xdgHome, nil
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homedir, ".cache"), nil
}

// ConfigHome returns the XDG user configuration home.
func ConfigHome() (string, error) {
	if xdgHome := os.Getenv("XDG_CONFIG_HOME"); xdgHome != "" {
		return xdgHome, nil
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homedir, ".config"), nil
}

// DataHome returns the XDG user data home.
func DataHome() (string, error) {
	if xdgHome := os.Getenv("XDG_DATA_HOME"); xdgHome != "" {
		return xdgHome, nil
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homedir, ".local", "share"), nil
}
