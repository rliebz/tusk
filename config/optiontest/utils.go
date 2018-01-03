package optiontest

import (
	"fmt"

	"github.com/rliebz/tusk/config/marshal"
	"github.com/rliebz/tusk/config/option"
	"github.com/rliebz/tusk/config/when"
)

// Create creates a custom option for testing purposes.
func Create(operators ...func(o *option.Option)) option.Option {
	o := option.Option{}

	for _, f := range operators {
		f(&o)
	}

	return o
}

// WithName is an operator that adds a name to an option.
func WithName(name string) func(o *option.Option) {
	return func(o *option.Option) {
		o.Name = name
	}
}

// WithDependency is an operator that adds a dependency to an option.
func WithDependency(name string) func(o *option.Option) {
	return func(o *option.Option) {
		o.DefaultValues = append(
			o.DefaultValues,
			option.Value{
				Value: fmt.Sprintf("${%s}", name),
			},
		)
	}
}

// WithWhenDependency is an operator that adds a when dependency to an option.
func WithWhenDependency(name string) func(o *option.Option) {
	return func(o *option.Option) {
		o.DefaultValues = append(
			o.DefaultValues,
			option.Value{
				When: when.When{
					Equal: map[string]marshal.StringList{
						name: {"true"},
					},
				},
			},
		)
	}
}
