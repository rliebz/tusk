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
	dirpath, err := os.Getwd()
	if err != nil {
		return "", false, err
	}

	for dirpath != "/" {
		fullPath, found, err = findFileInDir(dirpath)
		if err != nil || found {
			return fullPath, found, err
		}
		dirpath = filepath.Dir(dirpath)
	}

	return "", false, nil
}

func findFileInDir(dirpath string) (fullPath string, found bool, err error) {

	fullPath = filepath.Join(dirpath, DefaultFile)
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}

	return fullPath, true, nil
}
