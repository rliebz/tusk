package config

import (
	"fmt"

	"github.com/rliebz/tusk/config/when"
)

// createOption creates a custom option for testing purposes.
func createOption(operators ...func(o *Option)) Option {
	o := Option{}

	for _, f := range operators {
		f(&o)
	}

	return o
}

// withOptionName returns an operator that adds a name to an option.
func withOptionName(name string) func(o *Option) {
	return func(o *Option) {
		o.Name = name
	}
}

// withOptionDependency returns an operator that adds a dependency to an option.
func withOptionDependency(name string) func(o *Option) {
	return func(o *Option) {
		o.DefaultValues = append(
			o.DefaultValues,
			Value{
				Value: fmt.Sprintf("${%s}", name),
			},
		)
	}
}

// withOptionWhenDependency returns an operator that adds a when dependency to an option.
func withOptionWhenDependency(name string) func(o *Option) {
	return func(o *Option) {
		o.DefaultValues = append(
			o.DefaultValues,
			Value{When: when.List{when.Create(when.WithEqual(name, "true"))}},
		)
	}
}
