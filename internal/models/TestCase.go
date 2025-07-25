package models

type TestCaseInput struct {
	TestCaseId int    `json:"testCaseId"`
	Input      string `json:"input"`
}
type TestCaseOutput struct {
	TestCaseId int    `json:"testCaseId"`
	Output     string `json:"output"`
}
