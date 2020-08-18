package runner

import (
	"encoding/json"

	"github.com/rliebz/tusk/marshal"
)

// FindAllOptions returns a list of options relevant for a given config.
func FindAllOptions(t *Task, cfg *Config) ([]*Option, error) {
	names, err := getDependencies(t)
	if err != nil {
		return nil, err
	}

	candidates := make(map[string]*Option)
	for _, opt := range cfg.Options {
		// Args that share a name with global options take priority
		if _, ok := t.Args.Lookup(opt.Name); ok {
			continue
		}

		candidates[opt.Name] = opt
	}

	required := make([]*Option, 0, len(t.Options))
	for _, opt := range t.Options {
		candidates[opt.Name] = opt
		required = append(required, opt)
	}

	required, err = findRequiredOptionsRecursively(names, candidates, required)
	if err != nil {
		return nil, err
	}

	return required, nil
}

func findRequiredOptionsRecursively(
	entry []string, candidates map[string]*Option, found []*Option,
) ([]*Option, error) {
	for _, item := range entry {
		candidate, ok := candidates[item]
		if !ok || optionsContains(found, candidate) {
			continue
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

func optionsContains(items []*Option, item *Option) bool {
	for _, want := range items {
		if item == want {
			return true
		}
	}

	return false
}

type dependencyGetter interface {
	Dependencies() []string
}

func getDependencies(item dependencyGetter) ([]string, error) {
	// TODO: Remove json dependency by implementing stringer interface
	// json is used to print computed fields that should not be yaml parseable
	marshaled, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	names := marshal.FindPotentialVariables(marshaled)
	names = append(names, item.Dependencies()...)

	return names, nil
}
