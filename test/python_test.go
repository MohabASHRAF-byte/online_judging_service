package processorpackage

import (
	"judging-service/api/Dtos"
	"judging-service/containers"
	"judging-service/internal/models"
	"judging-service/internal/processor"
	"strings"
	"testing"
)

var manager2 = containers.NewContainersPoolManger(10)

func TestRunPythonWithTestcases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		submission      Dtos.SubmissionQueueDto
		expectErr       bool
		errContains     string
		expectedVerdict int
		customChecker   func(*testing.T, models.JudgingResult)
	}{
		{
			name: "Simple Hello World",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 1,
				Code:         "print(\"Hello World!\")",
				Language:     0,
				MemoryLimit:  256,
				TimeLimit:    1.0,
				InputTests: []models.TestCaseInput{
					{TestCaseId: 1, Input: ""},
				},
			},
			expectedVerdict: 0, // Accepted
			customChecker: func(t *testing.T, result models.JudgingResult) {
				if len(result.Outputs) != 1 {
					t.Fatalf("Expected 1 output, but got %d", len(result.Outputs))
				}
				if result.Outputs[0].Output != "Hello World!" {
					t.Errorf("Expected output 'Hello World!', but got: %s", result.Outputs[0].Output)
				}
			},
		},
		{
			name: "Input and Output",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 2,
				Code:         "name = input()\nprint(f\"Hello {name}!\")",
				Language:     0,
				MemoryLimit:  256,
				TimeLimit:    1.0,
				InputTests: []models.TestCaseInput{
					{TestCaseId: 1, Input: "World"},
					{TestCaseId: 2, Input: "Alice"},
				},
			},
			expectedVerdict: 0,
			customChecker: func(t *testing.T, result models.JudgingResult) {
				if len(result.Outputs) != 2 {
					t.Fatalf("Expected 2 outputs, but got %d", len(result.Outputs))
				}
				expectedOutputs := []string{"Hello World!", "Hello Alice!"}
				for i, output := range result.Outputs {
					if output.Output != expectedOutputs[i] {
						t.Errorf("Expected output '%s', but got: %s", expectedOutputs[i], output.Output)
					}
				}
			},
		},
		{
			name: "Math Operations",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 3,
				Code:         "a, b = map(int, input().split())\nprint(a + b)",
				Language:     0,
				MemoryLimit:  256,
				TimeLimit:    1.0,
				InputTests: []models.TestCaseInput{
					{TestCaseId: 1, Input: "3 4"},
					{TestCaseId: 2, Input: "10 20"},
				},
			},
			expectedVerdict: 0,
			customChecker: func(t *testing.T, result models.JudgingResult) {
				if len(result.Outputs) != 2 {
					t.Fatalf("Expected 2 outputs, but got %d", len(result.Outputs))
				}
				expectedOutputs := []string{"7", "30"}
				for i, output := range result.Outputs {
					if output.Output != expectedOutputs[i] {
						t.Errorf("Expected output '%s', but got: %s", expectedOutputs[i], output.Output)
					}
				}
			},
		},
		{
			name: "Syntax Error",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 4,
				Code:         "print(\"Hello World!\"\n# Missing closing parenthesis",
				Language:     0,
				MemoryLimit:  256,
				TimeLimit:    1.0,
				InputTests: []models.TestCaseInput{
					{TestCaseId: 1, Input: ""},
				},
			},
			expectErr:       true,
			errContains:     "compilation failed",
			expectedVerdict: 3, // Runtime Error (or compilation error)
		},
		{
			name: "Simple Script No Timeout",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 5,
				Code:         "# This is a simple script that should not timeout during syntax check\nprint('Hello')",
				Language:     0,
				MemoryLimit:  256,
				TimeLimit:    5.0,
				InputTests: []models.TestCaseInput{
					{TestCaseId: 1, Input: ""},
				},
			},
			expectedVerdict: 0,
			customChecker: func(t *testing.T, result models.JudgingResult) {
				if len(result.Outputs) != 1 {
					t.Fatalf("Expected 1 output, but got %d", len(result.Outputs))
				}
				if result.Outputs[0].Output != "Hello" {
					t.Errorf("Expected output 'Hello', but got: %s", result.Outputs[0].Output)
				}
			},
		},
		{
			name: "Runtime Time Limit Exceeded",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 6,
				Code:         "n = int(input())\nx = 0\nfor i in range(n):\n    x += 90 * 2 + i\nprint(x)",
				Language:     0,
				MemoryLimit:  256,
				TimeLimit:    1.0,
				InputTests: []models.TestCaseInput{
					{TestCaseId: 1, Input: "1000000000"},
				},
			},
			expectErr:       true,
			errContains:     "Time Limit",
			expectedVerdict: 2, // Time Limit Exceeded
		},
		{
			name: "Infinite Loop Time Limit",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 7,
				Code:         "while True:\n    pass",
				Language:     0,
				MemoryLimit:  256,
				TimeLimit:    1.0,
				InputTests: []models.TestCaseInput{
					{TestCaseId: 1, Input: ""},
				},
			},
			expectErr:       true,
			errContains:     "Time Limit",
			expectedVerdict: 2,
		},
		{
			name: "Empty Test Cases",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 8,
				Code:         "print(\"No input needed\")",
				Language:     0,
				MemoryLimit:  256,
				TimeLimit:    1.0,
				InputTests:   []models.TestCaseInput{}, // Empty test cases
			},
			expectedVerdict: 0,
			customChecker: func(t *testing.T, result models.JudgingResult) {
				if len(result.Outputs) != 0 {
					t.Errorf("Expected 0 outputs for empty test cases, but got %d", len(result.Outputs))
				}
			},
		},
		{
			name: "Large Output",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 9,
				Code:         "for i in range(1, 101):\n    print(i, end=' ')",
				Language:     0,
				MemoryLimit:  256,
				TimeLimit:    1.0,
				InputTests: []models.TestCaseInput{
					{TestCaseId: 1, Input: ""},
				},
			},
			expectedVerdict: 0,
			customChecker: func(t *testing.T, result models.JudgingResult) {
				if len(result.Outputs) != 1 {
					t.Fatalf("Expected 1 output, but got %d", len(result.Outputs))
				}
				output := result.Outputs[0].Output
				if !strings.Contains(output, "1 ") || !strings.Contains(output, "100") {
					t.Errorf("Expected output to contain '1 ' and '100', but got: %s", output)
				}
			},
		},
		{
			name: "Multiple Lines Input",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 10,
				Code:         "lines = []\nfor i in range(3):\n    lines.append(input())\nprint(' '.join(lines))",
				Language:     0,
				MemoryLimit:  256,
				TimeLimit:    1.0,
				InputTests: []models.TestCaseInput{
					{TestCaseId: 1, Input: "Hello\nWorld\nPython"},
				},
			},
			expectedVerdict: 0,
			customChecker: func(t *testing.T, result models.JudgingResult) {
				if len(result.Outputs) != 1 {
					t.Fatalf("Expected 1 output, but got %d", len(result.Outputs))
				}
				if result.Outputs[0].Output != "Hello World Python" {
					t.Errorf("Expected output 'Hello World Python', but got: %s", result.Outputs[0].Output)
				}
			},
		},
		{
			name: "List Processing",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 11,
				Code:         "numbers = list(map(int, input().split()))\nresult = sum(numbers)\nprint(result)",
				Language:     0,
				MemoryLimit:  256,
				TimeLimit:    1.0,
				InputTests: []models.TestCaseInput{
					{TestCaseId: 1, Input: "1 2 3 4 5"},
					{TestCaseId: 2, Input: "10 20 30"},
				},
			},
			expectedVerdict: 0,
			customChecker: func(t *testing.T, result models.JudgingResult) {
				if len(result.Outputs) != 2 {
					t.Fatalf("Expected 2 outputs, but got %d", len(result.Outputs))
				}
				expectedOutputs := []string{"15", "60"}
				for i, output := range result.Outputs {
					if output.Output != expectedOutputs[i] {
						t.Errorf("Expected output '%s', but got: %s", expectedOutputs[i], output.Output)
					}
				}
			},
		},
		{
			name: "String Operations",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 12,
				Code:         "text = input()\nprint(text.upper())\nprint(text.lower())\nprint(len(text))",
				Language:     0,
				MemoryLimit:  256,
				TimeLimit:    1.0,
				InputTests: []models.TestCaseInput{
					{TestCaseId: 1, Input: "Hello"},
				},
			},
			expectedVerdict: 0,
			customChecker: func(t *testing.T, result models.JudgingResult) {
				if len(result.Outputs) != 1 {
					t.Fatalf("Expected 1 output, but got %d", len(result.Outputs))
				}
				expected := "HELLO\nhello\n5"
				if result.Outputs[0].Output != expected {
					t.Errorf("Expected output '%s', but got: %s", expected, result.Outputs[0].Output)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := processor.RunCodeWithTestcases(manager2, tc.submission)

			if tc.expectErr {
				if err == nil {
					t.Fatalf("Expected an error, but got none")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("Expected error to contain '%s', but got: %v", tc.errContains, err)
				}
				if result.Verdict != tc.expectedVerdict {
					t.Errorf("Expected verdict %d, but got %d", tc.expectedVerdict, result.Verdict)
				}
				if !result.IsErrorExist {
					t.Errorf("Expected IsErrorExist to be true, but got false")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				if result.Verdict != tc.expectedVerdict {
					t.Errorf("Expected verdict %d, but got %d", tc.expectedVerdict, result.Verdict)
				}
				if result.IsErrorExist {
					t.Errorf("Expected IsErrorExist to be false, but got true")
				}
				if tc.customChecker != nil {
					tc.customChecker(t, result)
				}
			}
		})
	}
}
