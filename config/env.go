package config

import (
	"fmt"
)

// // replaceArgs evaluates the variables given and returns interpolated text.
// func replaceArgs(text []byte) ([]byte, error) {
// 	ordered := new(struct {
// 		Args yaml.MapSlice
// 	})

// 	if err := yaml.Unmarshal(text, ordered); err != nil {
// 		return nil, err
// 	}

// 	interpolatable, err := regexp.Compile(`{{\s*\w+\s*}}`)
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, mapslice := range ordered.Args {
// 		// Get ENV struct from initial text
// 		config := New()

// 		if err := yaml.Unmarshal(text, &config); err != nil {
// 			return nil, err
// 		}

// 		name, ok := mapslice.Key.(string)
// 		if !ok {
// 			return nil, fmt.Errorf("failed to assert name as string: %v", mapslice.Key)
// 		}

// 		arg, ok := config.Args[name]
// 		if !ok {
// 			return nil, fmt.Errorf("arg `%s` not found in config file args", name)
// 		}

// 		// TODO: CLI Flags
// 		// Attempt to get value from environment
// 		value := os.Getenv(arg.Environment)
// 		// Default to default value
// 		if value == "" {
// 			d := arg.Default
// 			value, ok = d.(string)
// 			if arg.Default != nil && !ok {
// 				return nil, fmt.Errorf("failed to assert default value as string: %v", d)
// 			}
// 		}

// 		// Warn about interpolation syntax in value
// 		if interpolatable.MatchString(value) {
// 			warning := fmt.Sprintf(
// 				"potential interpolation not evaluated: %s",
// 				value,
// 			)
// 			ui.Warn(warning)
// 		}

// 		// interpolate over remainder of config file
// 		pattern := InterpolationPattern(name)
// 		re, err := regexp.Compile(pattern)
// 		if err != nil {
// 			return nil, err
// 		}

// 		text = re.ReplaceAllLiteral(text, []byte(value))
// 	}

// 	return text, nil
// }

// InterpolationPattern returns the regexp pattern for a given name.
func InterpolationPattern(name string) string {
	return fmt.Sprintf("{{\\s*%s\\s*}}", name)
}
