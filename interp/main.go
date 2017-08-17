package interp

import (
	"bytes"
	"fmt"
	"regexp"
)

var escSeq = "{UNLIKELY_ESCAPE_SEQUENCE}"

// Interpolate runs the interpolation
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

// CompileGeneric returns the regexp pattern to identify a potential variable.
func CompileGeneric() *regexp.Regexp {
	return regexp.MustCompile(`\${(\w+)}`)
}

// Compile returns the regexp pattern for a given variable name.
func Compile(name string) (*regexp.Regexp, error) {
	pattern := fmt.Sprintf(`\$({%s})`, name)
	return regexp.Compile(pattern)
}

func escapePattern(text []byte) []byte {
	return bytes.Replace(text, []byte("$$"), []byte(escSeq), -1)
}

func unescapePattern(text []byte) []byte {
	return bytes.Replace(text, []byte(escSeq), []byte("$"), -1)
}
