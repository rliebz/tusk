package appcli

import (
	"os"
	"path/filepath"
)

var defaultFiles = []string{"tusk.yml", "tusk.yaml"}

// searchForFile checks the working directory and every parent directory to
// find a configuration file with the default name. If no file is found, an
// empty string will be returned
//
// This should be called when an explicit file is not passed in to determine
// the full path to the relevant config file.
func searchForFile() (string, error) {
	prevPath := ""
	dirPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for dirPath != prevPath {
		fullPath, found, err := findFileInDir(dirPath)
		if err != nil || found {
			return fullPath, err
		}
		prevPath, dirPath = dirPath, filepath.Dir(dirPath)
	}

	return "", nil
}

func findFileInDir(dirPath string) (fullPath string, found bool, err error) {
	for _, fileName := range defaultFiles {
		fullPath, found, err = findFileInDirByName(dirPath, fileName)
		if err != nil || found {
			return fullPath, found, err
		}
	}

	return "", false, nil
}

func findFileInDirByName(dirPath, fileName string) (fullPath string, found bool, err error) {
	fullPath = filepath.Join(dirPath, fileName)
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}

	return fullPath, true, nil
}
