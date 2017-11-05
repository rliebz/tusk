package config

import (
	"os"
	"path/filepath"
)

// DefaultFile is the default name for a config file.
var DefaultFile = "tusk.yml"

// SearchForFile checks the working directory and every parent directory to
// find a configuration file with the default name.
// This should be called when an explicit file is not passed in to determine
// the full path to the relevant config file.
func SearchForFile() (fullPath string, found bool, err error) {
	prevPath := ""
	dirPath, err := os.Getwd()
	if err != nil {
		return "", false, err
	}

	for dirPath != prevPath {
		fullPath, found, err = findFileInDir(dirPath)
		if err != nil || found {
			return fullPath, found, err
		}
		prevPath, dirPath = dirPath, filepath.Dir(dirPath)
	}

	return "", false, nil
}

func findFileInDir(dirPath string) (fullPath string, found bool, err error) {

	fullPath = filepath.Join(dirPath, DefaultFile)
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}

	return fullPath, true, nil
}
