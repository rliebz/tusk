package interp

import (
	"bytes"
	"fmt"
	"regexp"

	yaml "gopkg.in/yaml.v2"
)

var escSeq = []byte("{UNLIKELY_ESCAPE_SEQUENCE}")

// Marshallable interpolates an arbitrary YAML-marshallable interface.
func Marshallable(i interface{}, values map[string]string) error {
	text, err := yaml.Marshal(i)
	if err != nil {
		return err
	}

	text, err = mapInterpolate(text, values)
	if err != nil {
		return err
	}

	text = escape(text)

	return yaml.UnmarshalStrict(text, i)
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

// escape escapes all instances of $$ with $.
func escape(text []byte) []byte {
	return bytes.ReplaceAll(text, []byte("$$"), []byte("$"))
}

// interpolate replaces instances of the name pattern with the value.
func interpolate(text []byte, name, value string) ([]byte, error) {
	text = escapePattern(text)

	re, err := compile(name)
	if err != nil {
		return nil, err
	}

	text = re.ReplaceAll(text, []byte(value))

	return unescapePattern(text), nil
}

// mapInterpolate runs interpolation over a map from variable name to value.
func mapInterpolate(text []byte, m map[string]string) ([]byte, error) {

	for variable, value := range m {
		var err error
		text, err = interpolate(text, variable, value)
		if err != nil {
			return nil, err
		}
	}

	return text, nil
}

// compile returns the regexp pattern for a given variable name.
func compile(name string) (*regexp.Regexp, error) {
	pattern := fmt.Sprintf(`\$({%s})`, name)
	return regexp.Compile(pattern)
}

// escapePattern escapes unwanted potential interpolation targets.
func escapePattern(text []byte) []byte {
	return bytes.ReplaceAll(text, []byte("$$"), escSeq)
}

// unescapePattern returns unwanted potential interpolation targets.
func unescapePattern(text []byte) []byte {
	return bytes.ReplaceAll(text, escSeq, []byte("$$"))
}
