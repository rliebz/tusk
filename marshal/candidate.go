package marshal

import (
	"errors"

	yaml "gopkg.in/yaml.v2"
)

// UnmarshalCandidate is a candidate for unmarshaling.
// Candidates should only be defined inside of an UnmarshalYAML function.
type UnmarshalCandidate struct {
	// Unmarshal should return the result of UnmarshalYAML's unmarshal function.
	// This simply provides a closure so that different data types can be
	// safely passed into the unmarshaling function without reflection.
	Unmarshal func() error

	// Assign assigns the newly unmarshaled item using a closure.
	// This allows the resulting value from an unmarshaling to be assigned
	// to the receiver of the custom UnmarshalYAML function.
	Assign func()

	// Validate is an optional function that can validate after unmarshaling.
	// Assignment will not occur if validation fails.
	Validate func() error
}

// UnmarshalOneOf unmarshals candidates of different types until successful.
// If any error other than a yaml.TypeError is thrown, that error is returned
// immediately. If no candidates are valid, the error from the last candidate
// passed will be returned.
func UnmarshalOneOf(candidates ...UnmarshalCandidate) error {
	err := errors.New("no candidates passed")

	for _, c := range candidates {
		if err = c.Unmarshal(); err != nil {
			// TypeErrors are expected; try the next candidate
			if _, ok := err.(*yaml.TypeError); ok {
				continue
			}

			return err
		}

		if c.Validate != nil {
			if validationErr := c.Validate(); validationErr != nil {
				return validationErr
			}
		}

		if c.Assign != nil {
			c.Assign()
		}

		return nil
	}

	return err
}
