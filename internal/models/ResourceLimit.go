package models

type ResourceLimit struct {
	MemoryLimitInMB    int
	TimeLimitInSeconds float32
	CPU                int
}
