package option

import (
	"fmt"

	"github.com/rliebz/tusk/config/when"
)

// Create creates a custom option for testing purposes.
func Create(operators ...func(o *Option)) Option {
	o := Option{}

	for _, f := range operators {
		f(&o)
	}

	return o
}

// WithName returns an operator that adds a name to an option.
func WithName(name string) func(o *Option) {
	return func(o *Option) {
		o.Name = name
	}
}

// WithDependency returns an operator that adds a dependency to an option.
func WithDependency(name string) func(o *Option) {
	return func(o *Option) {
		o.DefaultValues = append(
			o.DefaultValues,
			Value{
				Value: fmt.Sprintf("${%s}", name),
			},
		)
	}
}

// WithWhenDependency returns an operator that adds a when dependency to an option.
func WithWhenDependency(name string) func(o *Option) {
	return func(o *Option) {
		o.DefaultValues = append(
			o.DefaultValues,
			Value{When: when.Create(when.WithEqual(name, "true"))},
		)
	}
}
