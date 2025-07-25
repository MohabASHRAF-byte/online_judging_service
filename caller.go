package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"judging-service/containers"
	"judging-service/internal/models"
	"judging-service/internal/processor"
	"log"
	"net/http"
	"time"
)

type SubmissionRequest struct {
	SubmissionId int                    `json:"submissionId"`
	Code         string                 `json:"code"`
	Language     int                    `json:"language"`
	MemoryLimit  int                    `json:"memoryLimit"`
	TimeLimit    float32                `json:"timeLimit"`
	InputTests   []models.TestCaseInput `json:"inputTests"`
}

type JudgmentResult struct {
	SubmissionId int      `json:"submissionId"`
	IsErrorExist bool     `json:"isErrorExist"`
	FallingTest  int      `json:"fallingTest"`
	Verdict      int      `json:"verdict"`
	Outputs      []string `json:"outputs"`
}

type JudgeService struct {
	baseURL string
	client  *http.Client
}

func NewJudgeService(baseURL string) *JudgeService {
	return &JudgeService{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (js *JudgeService) fetchSubmissions(limit int) ([]SubmissionRequest, error) {
	url := fmt.Sprintf("%s/api/SubmissionQueue?limit=%d", js.baseURL, limit)

	resp, err := js.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch submissions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var submissions []SubmissionRequest
	if err := json.Unmarshal(body, &submissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal submissions: %w", err)
	}

	return submissions, nil
}

func (js *JudgeService) submitResult(result JudgmentResult) error {
	url := fmt.Sprintf("%s/api/SubmissionQueue", js.baseURL)

	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	resp, err := js.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to submit result: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
func (js *JudgeService) processSubmissions() {
	submissions, err := js.fetchSubmissions(1)
	if err != nil {
		log.Printf("Error fetching submissions: %v", err)
		return
	}

	if len(submissions) == 0 {
		log.Println("No submissions to process")
		return
	}

	for _, submission := range submissions {
		log.Printf("Processing submission %d", submission.SubmissionId)
		var manger = containers.NewContainersPoolManger(10)
		var lang string
		if submission.Language == 1 {
			lang = string(models.Cpp)
		} else if submission.Language == 0 {
			lang = string(models.Python)
		}
		var output []models.TestCaseOutput
		output, _ = processor.RunCodeWithTestcases(
			manger,
			submission.Code,
			submission.InputTests,
			lang,
			models.ResourceLimit{MemoryLimitInMB: submission.MemoryLimit, TimeLimitInSeconds: submission.TimeLimit})
		log.Printf(output[0].Output)

		log.Printf("Successfully processed submission %d", submission.SubmissionId)
	}
}

func main() {
	judgeService := NewJudgeService("http://localhost:5129")

	log.Println("Judge microservice started")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			judgeService.processSubmissions()
		}
	}
}
