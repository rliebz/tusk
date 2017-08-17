package config

import (
	"gitlab.com/rliebz/tusk/interp"
	"gitlab.com/rliebz/tusk/task"
	yaml "gopkg.in/yaml.v2"
)

// FindAllFlags returns a list of flags relevant for a given task.
func (cfg *Config) FindAllFlags(t *task.Task) ([]*task.Arg, error) {
	names, err := getDependencies(t)
	if err != nil {
		return nil, err
	}

	candidates := make(map[string]*task.Arg)
	for name, arg := range cfg.Args {
		arg.Name = name
		candidates[name] = arg
	}

	var required []*task.Arg
	for name, arg := range t.Args {
		arg.Name = name
		required = append(required, arg)
		candidates[name] = arg
	}

	required, err = recurseDependencies(names, candidates, required)
	if err != nil {
		return nil, err
	}

	return required, nil
}

func recurseDependencies(
	entry []string, candidates map[string]*task.Arg, found []*task.Arg,
) ([]*task.Arg, error) {

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
