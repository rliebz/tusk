package runner

import (
	"errors"
	"fmt"
)

// IsFailedCondition checks if an error was because of a failed condition.
func IsFailedCondition(err error) bool {
	var cf conditionFailed
	return errors.As(err, &cf) && cf.WhenConditionFailed()
}

type conditionFailed interface {
	WhenConditionFailed() bool
}

type conditionFailedError struct {
	message string
}

func (e *conditionFailedError) Error() string {
	return e.message
}

func (e *conditionFailedError) WhenConditionFailed() bool {
	return true
}

// newCondFailError returns an error indicating a condition has failed.
func newCondFailError(msg string) error {
	return &conditionFailedError{msg}
}

// newCondFailErrorf returns an error indicating a condition has failed.
func newCondFailErrorf(msg string, a ...any) error {
	formatted := fmt.Sprintf(msg, a...)
	return &conditionFailedError{formatted}
}

// IsUnspecifiedClause checks if an error was because a clause is not defined.
func IsUnspecifiedClause(err error) bool {
	var uc unspecifiedClause
	return errors.As(err, &uc) && uc.WhenUnspecifiedClause()
}

type unspecifiedClauseError struct {
	message string
}

type unspecifiedClause interface {
	WhenUnspecifiedClause() bool
}

func (e *unspecifiedClauseError) Error() string {
	return e.message
}

func (e *unspecifiedClauseError) WhenUnspecifiedClause() bool {
	return true
}

// newUnspecifiedError returns an error for unspecified clauses.
func newUnspecifiedError(clauseName string) error {
	formatted := fmt.Sprintf("clause %q is not defined", clauseName)
	return &unspecifiedClauseError{formatted}
}
