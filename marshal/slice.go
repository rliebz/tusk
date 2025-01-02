package marshal

// Slice is a list of items optionally represented in yaml as a single item.
type Slice[T any] []T

// UnmarshalYAML unmarshals an item or list of items always into a list.
func (sl *Slice[T]) UnmarshalYAML(unmarshal func(any) error) error {
	var list []T
	listCandidate := UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&list) },
		Assign:    func() { *sl = list },
	}

	var single T
	singleCandidate := UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&single) },
		Assign:    func() { *sl = []T{single} },
	}

	return UnmarshalOneOf(listCandidate, singleCandidate)
}
