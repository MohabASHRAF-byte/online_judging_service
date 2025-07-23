package models

type Submission struct {
	ID         string   `json:"id"`
	Language   string   `json:"language"`
	Code       string   `json:"code"`
	TestInputs []string `json:"test_inputs"`
}

type ExecutionResult struct {
	ID      string   `json:"id"`
	Outputs []string `json:"outputs"`
	Status  string   `json:"status"`
}
