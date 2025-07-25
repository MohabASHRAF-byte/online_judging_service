package models

type JudgingResult struct {
	SubmissionId int              `json:"SubmissionId"`
	IsErrorExist bool             `json:"IsErrorExist"`
	FallingTest  int              `json:"FallingTest"`
	Verdict      int              `json:"Verdict"`
	Outputs      []TestCaseOutput `json:"Outputs"`
}
