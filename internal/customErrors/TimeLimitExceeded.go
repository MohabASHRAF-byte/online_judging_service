package customErrors

import (
	"fmt"
)

type TimeLimitExceededError struct {
	Operation string
	Limit     int
}

func (e *TimeLimitExceededError) Error() string {
	return fmt.Sprintf("Time Limit Exceeded after : %v s", e.Limit)
}
