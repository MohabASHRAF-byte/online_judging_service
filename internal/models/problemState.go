package models

type problemState string

const (
	Passed              problemState = "Passed"
	TimeLimitExceeded   problemState = "TimeLimitExceeded"
	MemoryLimitExceeded problemState = "MemoryLimitExceeded"
)
