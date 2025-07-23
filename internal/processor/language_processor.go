package processor

import (
	"fmt"
	"judging-service/containers"
	"judging-service/internal/models"
	"judging-service/internal/service"
	"time"
)

func RunCodeWithTestcases(m *containers.ContainersPoolManger, code string, testcases []string, codeLanguage string) ([]string, error) {

	var exec models.LangContainer
	var lang models.Language
	overallStart := time.Now()
	// switch for input language
	if string(models.Cpp) == codeLanguage {
		lang = models.Cpp
		exec = service.CppRunLangInterFace{}
	} else {
		return nil, fmt.Errorf("invalid language")
	}

	// init container
	doc, err := m.GetContainer(lang)
	defer m.FreeContainer(doc)

	// copy code to the container
	fileName, err := exec.CopyCodeToFile(doc, code)
	if err != nil {
		return nil, err
	}

	// compile code if compilable
	compileCommand, err := exec.CompileCode(doc, fileName)
	if err != nil {
		return nil, err
	}

	// run all test cases
	outputs, err := exec.RunTestCases(doc, testcases, compileCommand)
	if err != nil {
		return nil, err
	}
	fmt.Printf("🎯 Total Execution Time: %v\n", time.Since(overallStart))
	return outputs, nil
}

// ============================================
// HELPER FUNCTION: Create TAR Archive Directly from Memory
// ============================================
