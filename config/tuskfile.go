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

// Config is a struct representing the format for a Tuskfile.
type Config struct {
	Args  map[string]*task.Arg
	Tasks map[string]*task.Task
}

// New is the constructor for Config.
func New() *Config {
	return &Config{
		Args:  make(map[string]*task.Arg),
		Tasks: make(map[string]*task.Task),
	}
}

// ReadTuskfile parses the contents of a tusk file
func ReadTuskfile() (*Config, error) {
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
		// No Tuskfile is equivalent to an empty Tuskfile
		return New(), nil
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return parseTuskfile(data)
}

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

func parseTuskfile(text []byte) (*Config, error) {
	tuskfile := New()

	if err := yaml.Unmarshal(text, &tuskfile); err != nil {
		log.Printf("error: %v\n", err)
		return nil, err
	}

	return tuskfile, nil
}
