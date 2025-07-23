package customErrors

import "fmt"

type CompilationError struct {
	Operation string
	Limit     int
}

func (e *CompilationError) Error() string {
	return fmt.Sprintf("Compilation error")
}
