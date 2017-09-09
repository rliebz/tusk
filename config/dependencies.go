package config

import (
	"fmt"

	"github.com/rliebz/tusk/config/task"
	"github.com/rliebz/tusk/config/task/option"
	"github.com/rliebz/tusk/interp"
	yaml "gopkg.in/yaml.v2"
)

// AddSubTasks will recursively add task objects to the task's list of pretasks.
func AddSubTasks(cfg *Config, t *task.Task) error {

	for _, run := range t.Run {
		for _, subTaskName := range run.Task {
			// TODO: This requires tasks to be defined in order
			subTask, ok := cfg.Tasks[subTaskName]
			if !ok {
				return fmt.Errorf("sub-task %s was referenced before definition", subTaskName)
			}

			t.SubTasks = append(t.SubTasks, subTask)
			if err := AddSubTasks(cfg, subTask); err != nil {
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
		opt.Name = name
		candidates[name] = opt
	}

	var required []*option.Option
	for name, opt := range t.Options {
		opt.Name = name
		candidates[name] = opt
		required = append(required, opt)
	}

	required, err = recurseDependencies(names, candidates, required)
	if err != nil {
		return nil, err
	}

	for _, subTask := range t.SubTasks {
		nested, err := cfg.FindAllOptions(subTask)
		if err != nil {
			return nil, err
		}

		required = joinListsUnique(required, nested)
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

	re := interp.CompileGeneric()
	groups := re.FindAllStringSubmatch(string(marshalled), -1)

	var names []string
	for _, group := range groups {
		names = append(names, group[1])
	}

	names = append(names, item.Dependencies()...)

	return names, nil
}

func joinListsUnique(l1 []*option.Option, l2 []*option.Option) []*option.Option {
	set := make(map[*option.Option]struct{})
	for _, t := range l1 {
		set[t] = struct{}{}
	}
	for _, t := range l2 {
		set[t] = struct{}{}
	}

	var output []*option.Option
	for t := range set {
		output = append(output, t)
	}

	return output
}
