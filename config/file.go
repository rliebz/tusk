package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
)

// DefaultFile is the default name for a config file.
var DefaultFile = "tusk.yml"

// FindFile finds a config file and returns its contents.
// A blank filename can be passed, in which case a file will be searched for.
// Not finding a file is equivalent to finding an empty file.
func FindFile(filename string) ([]byte, error) {
	found := false
	passed := false

	if filename != "" {
		passed = true
	}

	if !passed {
		var err error
		filename, found, err = searchForFile()
		if err != nil {
			return nil, err
		}
	}

	if !passed && !found {
		return []byte{}, nil
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		message := fmt.Sprintf("failed to parse %s", filename)
		return nil, errors.Wrap(err, message)
	}

	return data, nil
}

func searchForFile() (filename string, found bool, err error) {
	dirpath, err := os.Getwd()
	if err != nil {
		return "", false, err
	}

	for dirpath != "/" {
		filename, found, err = findFileInDir(dirpath)
		if err != nil || found {
			return filename, found, err
		}
		dirpath = filepath.Dir(dirpath)
	}

	return "", false, nil
}

func findFileInDir(dirpath string) (filename string, found bool, err error) {

	filename = path.Join(dirpath, DefaultFile)
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}

	return filename, true, nil
}
