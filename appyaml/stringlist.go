package appyaml

import "errors"

// StringList is a list of strings optinally represented in yaml as a string.
// A single string in yaml will be unmarshalled as the first entry in a list,
// so the internal representation is always a list.
type StringList struct {
	Values []string
}

// UnmarshalYAML unmarshals a string or list of strings always into a list.
func (sl *StringList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var list []string
	if err := unmarshal(&list); err == nil {
		sl.Values = list
		return nil
	}

	var single string
	if err := unmarshal(&single); err == nil {
		sl.Values = []string{single}

		return nil
	}

	return errors.New("item is neither a string nor a list")
}
