package config

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"gitlab.com/rliebz/tusk/task"
	yaml "gopkg.in/yaml.v2"
)

// DefaultTuskfile is the default name for a Tuskfile.
var DefaultTuskfile = "tusk.yml"

func parseFileFlag(args []string) (tuskfile string, passed bool) {
	for i, arg := range args {
		if arg == "-f" || arg == "--file" {
			if i == len(args)-1 {
				// This error will be handled during cli.App#Run()
				return "", false
			}
			filename := args[i+1]
			return filename, true
		}
	}

	return "", false
}

func findTuskfile() (tuskfile string, found bool, err error) {
	dirpath, err := os.Getwd()
	if err != nil {
		return "", false, err
	}

	for dirpath != "/" {
		tuskfile, found, err = findTuskfileInDir(dirpath)
		if err != nil || found {
			return tuskfile, found, err
		}
		dirpath = filepath.Dir(dirpath)
	}

	return "", false, nil
}

func findTuskfileInDir(dirpath string) (tuskfile string, found bool, err error) {

	tuskfile = path.Join(dirpath, DefaultTuskfile)
	if _, err := os.Stat(tuskfile); err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}

	return tuskfile, true, nil
}

// ReadTuskfile parses the contents of a tusk file
func ReadTuskfile() (map[string]*task.Task, error) {
	tasks := make(map[string]*task.Task)
	found := false

	filename, passed := parseFileFlag(os.Args)

	if !passed {
		var err error
		filename, found, err = findTuskfile()
		if err != nil {
			return nil, err
		}
	}

	if !passed && !found {
		return tasks, nil
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &tasks)
	if err != nil {
		log.Printf("error: %v\n", err)
		return nil, err
	}

	return tasks, nil
}
