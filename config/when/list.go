package when

import "github.com/rliebz/tusk/config/marshal"

// List is a list of when items with custom yaml unmarshalling.
type List []When

// UnmarshalYAML allows single items to be used as lists.
func (l *List) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var whenSlice []When
	sliceCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&whenSlice) },
		Assign:    func() { *l = whenSlice },
	}

	var whenItem When
	itemCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&whenItem) },
		Assign:    func() { *l = List{whenItem} },
	}

	return marshal.UnmarshalOneOf(sliceCandidate, itemCandidate)
}

// Validate returns an error if any when clauses fail.
func (l *List) Validate(vars map[string]string) error {
	if l == nil {
		return nil
	}

	for _, w := range *l {
		if err := w.Validate(vars); err != nil {
			return err
		}
	}

	return nil
}

// Dependencies returns a list of options that are required explicitly.
// This does not include interpolations.
func (l *List) Dependencies() []string {
	if l == nil {
		return nil
	}

	// Use a map to prevent duplicates
	references := make(map[string]struct{})

	for _, w := range *l {
		for _, opt := range w.Dependencies() {
			references[opt] = struct{}{}
		}
	}

	options := make([]string, 0, len(references))
	for opt := range references {
		options = append(options, opt)
	}

	return options
}
