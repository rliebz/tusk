package marshal

// StringList is a list of strings optionally represented in yaml as a string.
// A single string in yaml will be unmarshalled as the first entry in a list,
// so the internal representation is always a list.
type StringList []string

// UnmarshalYAML unmarshals a string or list of strings always into a list.
func (sl *StringList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var single string
	singleCandidate := UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&single) },
		Assign:    func() { *sl = []string{single} },
	}

	var list []string
	listCandidate := UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&list) },
		Assign:    func() { *sl = list },
	}

	return UnmarshalOneOf(singleCandidate, listCandidate)
}
