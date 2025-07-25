package Dtos

import "judging-service/internal/models"

type SubmissionQueueDto struct {
	SubmissionId int                    `json:"submissionId"`
	Code         string                 `json:"code"`
	Language     int                    `json:"language"`
	MemoryLimit  int                    `json:"memoryLimit"`
	TimeLimit    float32                `json:"timeLimit"`
	InputTests   []models.TestCaseInput `json:"inputTests"`
}
