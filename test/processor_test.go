package processor_test

import (
	"judging-service/containers"
	"judging-service/internal/models"
	"judging-service/internal/processor"
	"reflect"
	"strings"
	"testing"
)

var manger = containers.NewContainersPoolManger(10)
var cpp = "cpp"

func TestRunCppWithTestcases(t *testing.T) {
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
			code:      "#include <iostream>\nint main() { std::cout << \"Hello World!\"; }",
			testcases: []string{""},
			resource:  defaultResource,
			expected:  []string{"Hello World!"},
			language:  models.Cpp,
		},
		{
			name:      "Input and Output",
			code:      "#include <iostream>\n#include <string>\nint main() { std::string name; std::cin >> name; std::cout << \"Hello \" << name << \"!\"; }",
			testcases: []string{"World", "Alice"},
			resource:  defaultResource,
			expected:  []string{"Hello World!", "Hello Alice!"},
			language:  models.Cpp,
		},
		{
			name:      "Math Operations",
			code:      "#include <iostream>\nint main() { int a, b; std::cin >> a >> b; std::cout << a + b; }",
			testcases: []string{"3 4", "10 20"},
			resource:  defaultResource,
			expected:  []string{"7", "30"},
			language:  models.Cpp,
		},
		{
			name:        "Compilation Error",
			code:        "#include <iostream>\nint main() { undeclared_variable = 5; }",
			testcases:   []string{""},
			resource:    defaultResource,
			expectErr:   true,
			errContains: "compilation failed",
			language:    models.Cpp,
		},
		{
			name:      "compile Time Limit Exceeded",
			code:      "#include <iostream>\nint main() { while(true); }",
			testcases: []string{""},
			resource: models.ResourceLimit{
				MemoryLimitInMB:    256,
				TimeLimitInSeconds: 5,
				CPU:                1,
			},
			expectErr:   true,
			errContains: "Time Limit",
			language:    models.Cpp,
		},
		{
			name:      "Runtime Time Limit Exceeded",
			code:      "#include <iostream>\nusing namespace std;\nint main() {\nlong long n;cin>>n;\nlong long x =0 ;\nfor(long long i =0 ;i<n ;i++)x+=90*2+i;cout<<x<<endl;   return 0;\n}",
			testcases: []string{"1000000000"},
			resource: models.ResourceLimit{
				MemoryLimitInMB:    256,
				TimeLimitInSeconds: 1,
				CPU:                1,
			},
			expectErr:   true,
			errContains: "Time Limit",
			language:    models.Cpp,
		},
		{
			name:      "Empty Test Cases",
			code:      "#include <iostream>\nint main() { std::cout << \"No input needed\"; }",
			testcases: []string{},
			resource:  defaultResource,
			expected:  []string{},
			language:  models.Cpp,
		},
		{
			name:      "Large Output",
			code:      "#include <iostream>\nint main() { for(int i = 1; i <= 100; i++) { std::cout << i << \" \"; } }",
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
			language: models.Cpp,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			outputs, err := processor.RunCodeWithTestcases(manger, tc.code, tc.testcases, string(tc.language), tc.resource)

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
