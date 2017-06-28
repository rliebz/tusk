package config

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"

	"gitlab.com/rliebz/tusk/task"
	yaml "gopkg.in/yaml.v2"
)

// DefaultTuskfile is the default name for a Tuskfile.
var DefaultTuskfile = "tusk.yml"

func parseFileFlag(args []string) (string, error) {
	for i, arg := range args {
		if arg == "-f" || arg == "--file" {
			if i == len(args)-1 {
				// This error will be handled during cli.App#Run()
				return "", nil
			}
			filename := args[i+1]
			return filename, nil
		}
	}

	return "", nil
}

func findTuskfile() (string, error) {
	// TODO: Search upwards through directories
	// TODO: Is no tuskfile an error?
	if _, err := os.Stat(DefaultTuskfile); os.IsNotExist(err) {
		return "", errors.Wrap(err, "Could not find a tuskfile")
	}

	return DefaultTuskfile, nil
}

// ReadTuskfile parses the contents of a tusk file
func ReadTuskfile() (map[string]*task.Task, error) {

	filename, err := parseFileFlag(os.Args)
	if err != nil {
		return nil, err
	}

	if filename == "" {
		filename, err = findTuskfile()
		if err != nil {
			return nil, err
		}
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	tasks := make(map[string]*task.Task)
	err = yaml.Unmarshal(data, &tasks)
	if err != nil {
		log.Printf("error: %v\n", err)
		return nil, err
	}

	return tasks, nil
}
