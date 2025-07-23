package processor

import (
	"judging-service/containers"
	"judging-service/internal/models"
	"judging-service/internal/processor"
	"reflect"
	"strings"
	"testing"
)

var manager = containers.NewContainersPoolManger(10)
var python = "python"

func TestRunPythonWithTestcases(t *testing.T) {
	t.Parallel()

	defaultResource := models.ResourceLimit{
		MemoryLimitInMB:    256,
		TimeLimitInSeconds: 1,
		CPU:                1,
	}

	testCases := []struct {
		name          string
		code          string
		testcases     []string
		resource      models.ResourceLimit
		expected      []string
		expectErr     bool
		errContains   string
		customChecker func(*testing.T, []string)
		language      models.Language
	}{
		{
			name:      "Simple Hello World",
			code:      "print(\"Hello World!\")",
			testcases: []string{""},
			resource:  defaultResource,
			expected:  []string{"Hello World!"},
			language:  models.Python,
		},
		{
			name:      "Input and Output",
			code:      "name = input()\nprint(f\"Hello {name}!\")",
			testcases: []string{"World", "Alice"},
			resource:  defaultResource,
			expected:  []string{"Hello World!", "Hello Alice!"},
			language:  models.Python,
		},
		{
			name:      "Math Operations",
			code:      "a, b = map(int, input().split())\nprint(a + b)",
			testcases: []string{"3 4", "10 20"},
			resource:  defaultResource,
			expected:  []string{"7", "30"},
			language:  models.Python,
		},
		{
			name:        "Syntax Error",
			code:        "print(\"Hello World!\"\n# Missing closing parenthesis",
			testcases:   []string{""},
			resource:    defaultResource,
			expectErr:   true,
			errContains: "compilation failed",
			language:    models.Python,
		},
		{
			name:      "Compile Time Limit Exceeded",
			code:      "# This is a simple script that should not timeout during syntax check\nprint('Hello')",
			testcases: []string{""},
			resource: models.ResourceLimit{
				MemoryLimitInMB:    256,
				TimeLimitInSeconds: 5,
				CPU:                1,
			},
			expected: []string{"Hello"},
			language: models.Python,
		},
		{
			name:      "Runtime Time Limit Exceeded",
			code:      "n = int(input())\nx = 0\nfor i in range(n):\n    x += 90 * 2 + i\nprint(x)",
			testcases: []string{"1000000000"},
			resource: models.ResourceLimit{
				MemoryLimitInMB:    256,
				TimeLimitInSeconds: 1,
				CPU:                1,
			},
			expectErr:   true,
			errContains: "Time Limit",
			language:    models.Python,
		},
		{
			name:      "Infinite Loop Time Limit",
			code:      "while True:\n    pass",
			testcases: []string{""},
			resource: models.ResourceLimit{
				MemoryLimitInMB:    256,
				TimeLimitInSeconds: 1,
				CPU:                1,
			},
			expectErr:   true,
			errContains: "Time Limit",
			language:    models.Python,
		},
		{
			name:      "Empty Test Cases",
			code:      "print(\"No input needed\")",
			testcases: []string{},
			resource:  defaultResource,
			expected:  []string{},
			language:  models.Python,
		},
		{
			name:      "Large Output",
			code:      "for i in range(1, 101):\n    print(i, end=' ')",
			testcases: []string{""},
			resource:  defaultResource,
			customChecker: func(t *testing.T, outputs []string) {
				if len(outputs) != 1 {
					t.Fatalf("Expected 1 output, but got %d", len(outputs))
				}
				if !strings.Contains(outputs[0], "1 ") || !strings.Contains(outputs[0], "100") {
					t.Errorf("Expected output to contain '1 ' and '100', but got: %s", outputs[0])
				}
			},
			language: models.Python,
		},
		{
			name:      "Multiple Lines Input",
			code:      "lines = []\nfor i in range(3):\n    lines.append(input())\nprint(' '.join(lines))",
			testcases: []string{"Hello\nWorld\nPython"},
			resource:  defaultResource,
			expected:  []string{"Hello World Python"},
			language:  models.Python,
		},
		{
			name:      "List Processing",
			code:      "numbers = list(map(int, input().split()))\nresult = sum(numbers)\nprint(result)",
			testcases: []string{"1 2 3 4 5", "10 20 30"},
			resource:  defaultResource,
			expected:  []string{"15", "60"},
			language:  models.Python,
		},
		{
			name:      "String Operations",
			code:      "text = input()\nprint(text.upper())\nprint(text.lower())\nprint(len(text))",
			testcases: []string{"Hello"},
			resource:  defaultResource,
			expected:  []string{"HELLO\nhello\n5"},
			language:  models.Python,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			outputs, err := processor.RunCodeWithTestcases(manager, tc.code, tc.testcases, string(tc.language), tc.resource)

			if tc.expectErr {
				if err == nil {
					t.Fatalf("Expected an error, but got none")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("Expected error to contain '%s', but got: %v", tc.errContains, err)
				}
				if outputs != nil && len(outputs) > 0 {
					t.Errorf("Expected no outputs due to error, but got: %v", outputs)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				if tc.customChecker != nil {
					tc.customChecker(t, outputs)
				} else if !reflect.DeepEqual(outputs, tc.expected) {
					t.Errorf("Expected outputs %v, but got %v", tc.expected, outputs)
				}
			}
		})
	}
}
