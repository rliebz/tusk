package interp

import (
	"bytes"
	"fmt"
	"regexp"
)

var escSeq = []byte("{UNLIKELY_ESCAPE_SEQUENCE}")

// Escape escapes all instances of $$ with $.
func Escape(text []byte) []byte {
	return bytes.Replace(text, []byte("$$"), []byte("$"), -1)
}

// Interpolate replaces instances of the name pattern with the value.
func Interpolate(text []byte, name string, value string) ([]byte, error) {
	text = escapePattern(text)

	re, err := Compile(name)
	if err != nil {
		return nil, err
	}

	text = re.ReplaceAll(text, []byte(value))

	return unescapePattern(text), nil
}

// Map runs interpolation over a map from variable name to value.
func Map(text []byte, m map[string]string) ([]byte, error) {

	for variable, value := range m {
		var err error
		text, err = Interpolate(text, variable, value)
		if err != nil {
			return nil, err
		}
	}

	return text, nil
}

// Contains verifies whether an interpolation string exists for a given name.
func Contains(text []byte, name string) (bool, error) {
	text = escapePattern(text)

	re, err := Compile(name)
	if err != nil {
		return false, err
	}

	return re.Match(text), nil
}

// FindPotentialVariables returns a list of potential interpolation target names.
func FindPotentialVariables(text []byte) []string {
	re := regexp.MustCompile(`\${([\w-]+)}`)

	groups := re.FindAllStringSubmatch(string(text), -1)

	names := make([]string, 0, len(groups))
	for _, group := range groups {
		names = append(names, group[1])
	}

	return names
}

// Compile returns the regexp pattern for a given variable name.
func Compile(name string) (*regexp.Regexp, error) {
	pattern := fmt.Sprintf(`\$({%s})`, name)
	return regexp.Compile(pattern)
}

// escapePattern escapes unwanted potential interpolation targets.
func escapePattern(text []byte) []byte {
	return bytes.Replace(text, []byte("$$"), escSeq, -1)
}

// unescapePattern returns unwanted potential interpolation targets.
func unescapePattern(text []byte) []byte {
	return bytes.Replace(text, escSeq, []byte("$$"), -1)
}
