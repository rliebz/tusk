package config

import (
	"github.com/rliebz/tusk/config/option"
	"github.com/rliebz/tusk/config/task"
	"github.com/rliebz/tusk/interp"
	yaml "gopkg.in/yaml.v2"
)

// FindAllOptions returns a list of options relevant for a given task.
func FindAllOptions(t *task.Task, cfg *Config) ([]*option.Option, error) {
	names, err := getDependencies(t)
	if err != nil {
		return nil, err
	}

	candidates := make(map[string]*option.Option)
	for name, opt := range cfg.Options {
		candidates[name] = opt
	}

	var required []*option.Option
	for name, opt := range t.Options {
		candidates[name] = opt
		required = append(required, opt)
	}

	required, err = findRequiredOptionsRecursively(names, candidates, required)
	if err != nil {
		return nil, err
	}

	return required, nil
}

func findRequiredOptionsRecursively(
	entry []string, candidates map[string]*option.Option, found []*option.Option,
) ([]*option.Option, error) {

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
		var dependencies []string
		for _, opt := range found {
			nested, err := getDependencies(opt)
			if err != nil {
				return nil, err
			}
			dependencies = append(dependencies, nested...)
		}

		var err error
		found, err = findRequiredOptionsRecursively(dependencies, candidates, found)
		if err != nil {
			return nil, err
		}
	}

	return found, nil
}

type dependencyGetter interface {
	Dependencies() []string
}

func getDependencies(item dependencyGetter) ([]string, error) {

	marshalled, err := yaml.Marshal(item)
	if err != nil {
		return nil, err
	}

	names := interp.FindPotentialVariables(marshalled)
	names = append(names, item.Dependencies()...)

	return names, nil
}
