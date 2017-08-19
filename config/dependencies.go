package config

import (
	"gitlab.com/rliebz/tusk/interp"
	"gitlab.com/rliebz/tusk/task"
	yaml "gopkg.in/yaml.v2"
)

// FindAllOptions returns a list of flags relevant for a given task.
func (cfg *Config) FindAllOptions(t *task.Task) ([]*task.Option, error) {
	names, err := getDependencies(t)
	if err != nil {
		return nil, err
	}

	candidates := make(map[string]*task.Option)
	for name, opt := range cfg.Options {
		opt.Name = name
		candidates[name] = opt
	}

	var required []*task.Option
	for name, opt := range t.Options {
		opt.Name = name
		required = append(required, opt)
		candidates[name] = opt
	}

	required, err = recurseDependencies(names, candidates, required)
	if err != nil {
		return nil, err
	}

	return required, nil
}

func recurseDependencies(
	entry []string, candidates map[string]*task.Option, found []*task.Option) ([]*task.Option, error) {

candidates:
	for _, item := range entry {
		candidate := candidates[item]

		if candidate == nil {
			continue
		}

		for _, f := range found {
			if f == candidate {
				continue candidates
			}
		}

		found = append(found, candidate)
		dependencies, err := getDependencies(found)
		if err != nil {
			return nil, err
		}

		found, err = recurseDependencies(dependencies, candidates, found)
		if err != nil {
			return nil, err
		}
	}

	return found, nil
}

func getDependencies(item interface{}) ([]string, error) {

	marshalled, err := yaml.Marshal(item)
	if err != nil {
		return nil, err
	}

	re := interp.CompileGeneric()
	groups := re.FindAllStringSubmatch(string(marshalled), -1)

	var names []string
	for _, group := range groups {
		names = append(names, group[1])
	}

	return names, nil
}
