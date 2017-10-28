package marshal

import yaml "gopkg.in/yaml.v2"

// Candidate is a candidate for unmarshalling.
// Candidates should only be defined inside of an UnmarshalYAML function.
type Candidate struct {
	// Unmarshal should return the result of UnmarshalYAML's unmarshal function.
	// This simply provides a closure so that different data types can be
	// safely passed into the unmarshalling function without reflection.
	Unmarshal func() error

	// Assign assigns the newly unmarshalled item using a closure.
	// This allows the resulting value from an unmarshalling to be assigned
	// to the receiver of the custom UnmarshalYAML function.
	Assign func()

	// Validate is an optional function that can validate after unmarshalling.
	// Assignment will not occur if validation fails.
	Validate func() error
}

// OneOf tries to unmarshal candidates of different types until successful.
// If any error other than a yaml.TypeError is thrown, that error is returned
// immediately. If no candidates are valid, the error from the last candidate
// passed will be returned.
func OneOf(candidates ...Candidate) error {
	var err error

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
