package marshal

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

// ParseOrderedMap is.
func ParseOrderedMap(
	ms yaml.MapSlice,
	assign func(string, []byte) error,
) ([]string, error) {
	ordered := make([]string, 0, len(ms))

	for _, itemMS := range ms {
		name, ok := itemMS.Key.(string)
		if !ok {
			return nil, fmt.Errorf("%q is not a valid key name", name)
		}
		ordered = append(ordered, name)

		text, err := yaml.Marshal(itemMS.Value)
		if err != nil {
			return nil, err
		}

		if err := assign(name, text); err != nil {
			return nil, err
		}
	}

	return ordered, nil
}
