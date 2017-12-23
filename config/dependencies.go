package config

import (
	"fmt"

	"github.com/rliebz/tusk/config/option"
	"github.com/rliebz/tusk/config/task"
	"github.com/rliebz/tusk/interp"
	yaml "gopkg.in/yaml.v2"
)

func addSubTasks(cfg *Config, cfgText []byte, t *task.Task) error {

	for _, run := range t.RunList {
		for _, subTaskDesc := range run.SubTaskList {
			st, ok := cfg.Tasks[subTaskDesc.Name]
			if !ok {
				return fmt.Errorf(
					`sub-task "%s" does not exist`,
					subTaskDesc.Name,
				)
			}

			passed := subTaskDesc.Options
			subTask := *st

			values, err := interpolateGlobalOptions(cfg, cfgText, passed, &subTask)
			if err != nil {
				return err
			}

			if err := interpolateTask(cfgText, values, passed, &subTask); err != nil {
				return err
			}

			run.Tasks = append(run.Tasks, subTask)

			if err := addSubTasks(cfg, cfgText, &subTask); err != nil {
				return err
			}

		}
	}

	return nil
}

// FindAllOptions returns a list of options relevant for a given task.
func (cfg *Config) FindAllOptions(t *task.Task) ([]*option.Option, error) {
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

	required, err = recurseDependencies(names, candidates, required)
	if err != nil {
		return nil, err
	}

	for _, run := range t.RunList {
		for _, subTask := range run.Tasks {
			nested, err := cfg.FindAllOptions(&subTask)
			if err != nil {
				return nil, err
			}

			required, err = addNestedDependencies(required, nested)
			if err != nil {
				return nil, err
			}
		}
	}

	return required, nil
}

func recurseDependencies(
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
		found, err = recurseDependencies(dependencies, candidates, found)
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

func addNestedDependencies(dependencies, nested []*option.Option) ([]*option.Option, error) {
	set := make(map[string]*option.Option)
	for _, opt := range dependencies {
		set[opt.Name] = opt
	}
	for _, newOpt := range nested {
		if found, ok := set[newOpt.Name]; ok {
			if newOpt != found {
				return nil, fmt.Errorf(
					`cannot redefine option "%s" in sub-task`, newOpt.Name,
				)
			}
			continue
		}

		dependencies = append(dependencies, newOpt)
	}

	return dependencies, nil
}
