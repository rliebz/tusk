package interp

import (
	"fmt"
	"regexp"
)

// Map runs interpolation over a map from variable name to value.
func Map(text []byte, m map[string]string) ([]byte, error) {

	for variable, value := range m {
		pattern := Pattern(variable)
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}

		text = re.ReplaceAll(text, []byte(value))
	}

	return text, nil
}

// Pattern returns the regexp pattern for a given name.
func Pattern(name string) string {
	return fmt.Sprintf("\\${\\s*%s\\s*}", name)
}
