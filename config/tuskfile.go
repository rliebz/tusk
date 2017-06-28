package config

import (
	"io/ioutil"
	"log"

	"gitlab.com/rliebz/tusk/task"
	yaml "gopkg.in/yaml.v2"
)

func ReadTuskfile(filename string) (map[string]*task.Task, error) {

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
