package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"judging-service/internal/models"
	"log"
	"net/http"
	"time"
)

type JudgeSubmissionRequest struct {
	SubmissionId int                       `json:"submissionId"`
	IsErrorExist bool                      `json:"isErrorExist"`
	FallingTest  *int                      `json:"fallingTest"`
	Verdict      int                       `json:"verdict"`
	Outputs      []JudgeProblemTestcaseDto `json:"outputs"`
}

type JudgeProblemTestcaseDto struct {
	TestCaseId int    `json:"testCaseId"`
	Output     string `json:"Output"`
}

func SubmitJudgingResult(result models.JudgingResult, baseURL string) error {
	url := fmt.Sprintf("%sapi/SubmissionQueue", baseURL)

	outputs := make([]JudgeProblemTestcaseDto, len(result.Outputs))
	for i, output := range result.Outputs {
		outputs[i] = JudgeProblemTestcaseDto{
			TestCaseId: output.TestCaseId,
			Output:     output.Output,
		}
	}

	var fallingTest *int
	if result.FallingTest > 0 {
		fallingTest = &result.FallingTest
	}

	request := JudgeSubmissionRequest{
		SubmissionId: result.SubmissionId,
		IsErrorExist: result.IsErrorExist,
		FallingTest:  fallingTest,
		Verdict:      result.Verdict,
		Outputs:      outputs,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to submit judging result: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Printf("Successfully submitted judging result for submission %d", result.SubmissionId)
	return nil
}
