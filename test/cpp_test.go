package processorpackage

import (
	"judging-service/api/Dtos"
	"judging-service/containers"
	"judging-service/internal/models"
	"judging-service/internal/processor"
	"strings"
	"testing"
)

var manager = containers.NewContainersPoolManger(10)

func TestRunCodeWithTestcases(t *testing.T) {
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
				Code:         "#include <iostream>\nint main() { std::cout << \"Hello World!\"; }",
				Language:     1, // Assuming 1 is C++
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
				Code:         "#include <iostream>\n#include <string>\nint main() { std::string name; std::cin >> name; std::cout << \"Hello \" << name << \"!\"; }",
				Language:     1,
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
				Code:         "#include <iostream>\nint main() { int a, b; std::cin >> a >> b; std::cout << a + b; }",
				Language:     1,
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
			name: "Compilation Error",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 4,
				Code:         "#include <iostream>\nint main() { undeclared_variable = 5; }",
				Language:     1,
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
			name: "Time Limit Exceeded",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 5,
				Code:         "#include <iostream>\nint main() { while(true); }",
				Language:     1,
				MemoryLimit:  256,
				TimeLimit:    1.0,
				InputTests: []models.TestCaseInput{
					{TestCaseId: 1, Input: ""},
				},
			},
			expectErr:       true,
			errContains:     "Time Limit",
			expectedVerdict: 2, // Time Limit Exceeded
		},
		{
			name: "Runtime Time Limit Exceeded with Large Input",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 6,
				Code:         "#include <iostream>\nusing namespace std;\nint main() {\nlong long n;cin>>n;\nlong long x =0 ;\nfor(long long i =0 ;i<n ;i++)x+=90*2+i;cout<<x<<endl;   return 0;\n}",
				Language:     1,
				MemoryLimit:  256,
				TimeLimit:    1.0,
				InputTests: []models.TestCaseInput{
					{TestCaseId: 1, Input: "1000000000"},
				},
			},
			expectErr:       true,
			errContains:     "Time Limit",
			expectedVerdict: 2,
		},
		{
			name: "Empty Test Cases",
			submission: Dtos.SubmissionQueueDto{
				SubmissionId: 7,
				Code:         "#include <iostream>\nint main() { std::cout << \"No input needed\"; }",
				Language:     1,
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
				SubmissionId: 8,
				Code:         "#include <iostream>\nint main() { for(int i = 1; i <= 100; i++) { std::cout << i << \" \"; } }",
				Language:     1,
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := processor.RunCodeWithTestcases(manager, tc.submission)

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
