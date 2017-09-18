package when

import "fmt"

// IsFailedCondition checks if an error was because of a failed condition.
func IsFailedCondition(err error) bool {
	we, ok := err.(conditionFailed)
	return ok && we.WhenConditionFailed()
}

type conditionFailed interface {
	WhenConditionFailed() bool
}

type conditionFailedError struct {
	Message string
}

func (e *conditionFailedError) Error() string {
	return e.Message
}

func (e *conditionFailedError) WhenConditionFailed() bool {
	return true
}

// newCondFailErrorf returns an error indicating a condition has failed.
func newCondFailErrorf(msg string, a ...interface{}) error {
	formatted := fmt.Sprintf(msg, a...)
	return &conditionFailedError{Message: formatted}
}
