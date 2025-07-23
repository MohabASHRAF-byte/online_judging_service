package processor

import (
	"context"
	"errors"
	"fmt"
	"judging-service/containers"
	"judging-service/internal/customErrors"
	"judging-service/internal/models"
	"judging-service/internal/service"
	"log"
	"time"
)

func RunCodeWithTestcases(m *containers.ContainersPoolManger, code string, testcases []string, codeLanguage string) ([]string, error) {
	var outputs []string
	for _, testCase := range testcases {
		testOutput, err := RuntestCase(m, code, testCase, codeLanguage)
		if err != nil {
			return nil, err
		}
		if testOutput != nil {
			outputs = append(outputs, *testOutput)
		}
	}
	return outputs, nil
}
func RuntestCase(m *containers.ContainersPoolManger, code string, testcase string, codeLanguage string) (*string, error) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Pass the context to the function you want to time out.
	compileCommand, err := exec.CompileCode(doc, fileName, ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, &customErrors.TimeLimitExceededError{
				Limit: 1,
			}
		}
		return nil, fmt.Errorf("compilation failed: %w", err)
	}
	log.Println("[DEBUG: 12] Compilation successful")
	// run all test cases
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Println("[DEBUG: 13] Compilation successful")

	output, err := exec.RunTestCases(doc, testcase, compileCommand, ctx)
	if errors.Is(err, context.DeadlineExceeded) {
		return nil, &customErrors.TimeLimitExceededError{}
	}
	if err != nil {
		log.Println("[DEBUG:14] Compilation successful")
		return nil, err
	}
	log.Println("[DEBUG: 13] Compilation successful")

	fmt.Printf("🎯 Total Execution Time: %v\n", time.Since(overallStart))
	return &output, nil
}
