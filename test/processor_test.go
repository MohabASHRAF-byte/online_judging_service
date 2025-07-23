package processor

import (
	"judging-service/containers"
	"judging-service/internal/processor"
	"strings"
	"testing"
)

var manger = containers.NewContainersPoolManger(10)
var cpp = "cpp"

func TestRunCppWithTestcases_SimpleHelloWorld(t *testing.T) {
	t.Parallel()
	code := `#include <iostream>
using namespace std;

int main() {
    cout << "Hello World!" << endl;
    return 0;
}`

	testcases := []string{""}
	expected := []string{"Hello World!"}

	outputs, err := processor.RunCodeWithTestcases(manger, code, testcases, cpp)

	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if len(outputs) != len(expected) {
		t.Fatalf("Expected %d outputs, but got %d", len(expected), len(outputs))
	}

	for i, output := range outputs {
		if output != expected[i] {
			t.Errorf("Test case %d: expected '%s', but got '%s'", i+1, expected[i], output)
		}
	}
}

func TestRunCppWithTestcases_InputOutput(t *testing.T) {
	t.Parallel()
	code := `#include <iostream>
#include <string>
using namespace std;

int main() {
    string name;
    cin >> name;
    cout << "Hello " << name << "!" << endl;
    return 0;
}`

	testcases := []string{"World", "Alice", "Bob"}
	expected := []string{"Hello World!", "Hello Alice!", "Hello Bob!"}

	outputs, err := processor.RunCodeWithTestcases(manger, code, testcases, cpp)

	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if len(outputs) != len(expected) {
		t.Fatalf("Expected %d outputs, but got %d", len(expected), len(outputs))
	}

	for i, output := range outputs {
		if output != expected[i] {
			t.Errorf("Test case %d: expected '%s', but got '%s'", i+1, expected[i], output)
		}
	}
}

func TestRunCppWithTestcases_MathOperations(t *testing.T) {
	t.Parallel()
	code := `#include <iostream>
using namespace std;

int main() {
    int a, b;
    cin >> a >> b;
    cout << a + b << endl;
    return 0;
}`

	testcases := []string{"3 4", "10 20", "100 50"}
	expected := []string{"7", "30", "150"}

	outputs, err := processor.RunCodeWithTestcases(manger, code, testcases, cpp)

	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if len(outputs) != len(expected) {
		t.Fatalf("Expected %d outputs, but got %d", len(expected), len(outputs))
	}

	for i, output := range outputs {
		if output != expected[i] {
			t.Errorf("Test case %d: expected '%s', but got '%s'", i+1, expected[i], output)
		}
	}
}

func TestRunCppWithTestcases_CompilationError(t *testing.T) {
	// Invalid C++ code that should fail to compile
	t.Parallel()

	code := `#include <iostream>
using namespace std;

int main() {
    undeclared_variable = 5;  // This should cause compilation error
    cout << "This won't work" << endl;
    return 0;
}`

	testcases := []string{""}

	outputs, err := processor.RunCodeWithTestcases(manger, code, testcases, cpp)

	// We expect an error due to compilation failure
	if err == nil {
		t.Fatalf("Expected compilation error, but got no error")
	}

	if !strings.Contains(err.Error(), "compilation failed") {
		t.Errorf("Expected compilation error message, but got: %v", err)
	}

	// Should have no outputs since compilation failed
	if outputs != nil && len(outputs) > 0 {
		t.Errorf("Expected no outputs due to compilation error, but got: %v", outputs)
	}
}

func TestRunCppWithTestcases_MultipleInputsPerCase(t *testing.T) {
	t.Parallel()

	code := `#include <iostream>
using namespace std;

int main() {
    int n;
    cin >> n;
    for(int i = 1; i <= n; i++) {
        cout << i << " ";
    }
    cout << endl;
    return 0;
}`

	testcases := []string{"3", "5", "1"}
	expected := []string{"1 2 3", "1 2 3 4 5", "1"}

	outputs, err := processor.RunCodeWithTestcases(manger, code, testcases, cpp)

	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if len(outputs) != len(expected) {
		t.Fatalf("Expected %d outputs, but got %d", len(expected), len(outputs))
	}

	for i, output := range outputs {
		if strings.TrimSpace(output) != expected[i] {
			t.Errorf("Test case %d: expected '%s', but got '%s'", i+1, expected[i], strings.TrimSpace(output))
		}
	}
}

func TestRunCppWithTestcases_EmptyTestCases(t *testing.T) {
	t.Parallel()

	code := `#include <iostream>
using namespace std;

int main() {
    cout << "No input needed" << endl;
    return 0;
}`

	testcases := []string{} // Empty test cases

	outputs, err := processor.RunCodeWithTestcases(manger, code, testcases, cpp)

	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if len(outputs) != 0 {
		t.Errorf("Expected 0 outputs for empty test cases, but got %d", len(outputs))
	}
}

func TestRunCppWithTestcases_LargeOutput(t *testing.T) {
	t.Parallel()

	code := `#include <iostream>
using namespace std;

int main() {
    for(int i = 1; i <= 100; i++) {
        cout << i << " ";
    }
    cout << endl;
    return 0;
}`

	testcases := []string{""}

	outputs, err := processor.RunCodeWithTestcases(manger, code, testcases, cpp)

	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if len(outputs) != 1 {
		t.Fatalf("Expected 1 output, but got %d", len(outputs))
	}

	// Check that output contains numbers from 1 to 100
	output := outputs[0]
	if !strings.Contains(output, "1 ") || !strings.Contains(output, "100") {
		t.Errorf("Expected output to contain '1 ' and '100 ', but got: %s", output)
	}
}
