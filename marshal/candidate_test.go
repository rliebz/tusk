package marshal

import (
	"errors"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	yaml "gopkg.in/yaml.v2"
)

func createTypeErrorCandidate(t *testing.T) UnmarshalCandidate {
	return UnmarshalCandidate{
		Unmarshal: func() error { return &yaml.TypeError{Errors: []string{"oh no"}} },
		Assign: func() {
			t.Error("unexpected call to Assign function")
		},
		Validate: func() error {
			t.Errorf("unexpected call to Validate function")
			return nil
		},
	}
}

func createFailCandidate(t *testing.T) UnmarshalCandidate {
	return UnmarshalCandidate{
		Unmarshal: func() error { return errors.New("oops") },
		Assign: func() {
			t.Error("unexpected call to Assign function")
		},
		Validate: func() error {
			t.Errorf("unexpected call to Validate function")
			return nil
		},
	}
}

func createInvalidCandidate(t *testing.T) UnmarshalCandidate {
	return UnmarshalCandidate{
		Unmarshal: func() error { return nil },
		Assign:    func() { t.Error("unexpected call to Assign function") },
		Validate:  func() error { return errors.New("expected failure") },
	}
}

func createNeverReachedCandidate(t *testing.T) UnmarshalCandidate {
	return UnmarshalCandidate{
		Unmarshal: func() error {
			t.Error("unexpected call to Unmarshal function")
			return nil
		},
	}
}

func TestOneOf_error(t *testing.T) {
	g := ghost.New(t)

	err := UnmarshalOneOf()
	g.Should(be.ErrorEqual(err, "no candidates passed"))

	err = UnmarshalOneOf(createTypeErrorCandidate(t))
	g.Should(be.Error(err))

	err = UnmarshalOneOf(
		createFailCandidate(t),
		createNeverReachedCandidate(t),
	)
	g.Should(be.Error(err))

	err = UnmarshalOneOf(
		createInvalidCandidate(t),
		createNeverReachedCandidate(t),
	)
	g.Should(be.Error(err))
}

func TestOneOf_success(t *testing.T) {
	g := ghost.New(t)

	validateCalled := false
	assignCalled := false

	t.Cleanup(func() {
		g.Check(validateCalled)
		g.Check(assignCalled)
	})

	successCandidate := UnmarshalCandidate{
		Unmarshal: func() error { return nil },
		Assign:    func() { assignCalled = true },
		Validate: func() error {
			validateCalled = true
			return nil
		},
	}

	err := UnmarshalOneOf(
		createTypeErrorCandidate(t),
		successCandidate,
		createNeverReachedCandidate(t),
	)
	g.NoError(err)
}
