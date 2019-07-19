package marshal

import (
	"errors"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func createTypeErrorCandidate(t *testing.T) UnmarshalCandidate {
	return UnmarshalCandidate{
		Unmarshal: func() error { return &yaml.TypeError{} },
		Assign: func() {
			t.Error("failed candidate called Assign function")
		},
		Validate: func() error {
			t.Errorf("failed candidate called Validate function")
			return nil
		},
	}
}

func createFailCandidate(t *testing.T) UnmarshalCandidate {
	return UnmarshalCandidate{
		Unmarshal: func() error { return errors.New("oops") },
		Assign: func() {
			t.Error("failed candidate called Assign function")
		},
		Validate: func() error {
			t.Errorf("failed candidate called Validate function")
			return nil
		},
	}
}

func createInvalidCandidate(t *testing.T) UnmarshalCandidate {
	return UnmarshalCandidate{
		Unmarshal: func() error { return nil },
		Assign:    func() { t.Error("invalid candidate called Assign function") },
		Validate:  func() error { return errors.New("expected failure") },
	}
}

func createNeverReachedCandidate(t *testing.T) UnmarshalCandidate {
	return UnmarshalCandidate{
		Unmarshal: func() error {
			t.Error("candidate was unexpectedly reached")
			return nil
		},
	}
}

func TestOneOf_error(t *testing.T) {
	if err := UnmarshalOneOf(); err == nil {
		t.Error("OneOf(): expected error, got nil")
	}

	if err := UnmarshalOneOf(
		createTypeErrorCandidate(t),
	); err == nil {
		t.Error("OneOf(typeError): expected error, got nil")
	}

	if err := UnmarshalOneOf(
		createFailCandidate(t),
		createNeverReachedCandidate(t),
	); err == nil {
		t.Error("OneOf(failed): expected error, got nil")
	}

	if err := UnmarshalOneOf(
		createInvalidCandidate(t),
		createNeverReachedCandidate(t),
	); err == nil {
		t.Error("OneOf(invalid): expected error, got nil")
	}
}

func TestOneOf_success(t *testing.T) {
	validateCalled := false
	assignCalled := false

	defer func() {
		if !validateCalled {
			t.Error(
				"OneOf(typeError, success): validate was never called on successful candidate",
			)
		}
		if !assignCalled {
			t.Error(
				"OneOf(typeError, success): assign was never called on successful candidate",
			)
		}
	}()

	successCandidate := UnmarshalCandidate{
		Unmarshal: func() error { return nil },
		Assign:    func() { assignCalled = true },
		Validate: func() error {
			validateCalled = true
			return nil
		},
	}

	if err := UnmarshalOneOf(
		createTypeErrorCandidate(t),
		successCandidate,
		createNeverReachedCandidate(t),
	); err != nil {
		t.Errorf(`OneOf(failed, invalid, success): unexpected error: %s`, err)
	}
}
