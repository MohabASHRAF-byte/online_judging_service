package processor

import (
	"context"
	"errors"
	"fmt"
	"judging-service/containers"
	"judging-service/internal/customErrors"
	"judging-service/internal/models"
	"log"
	"time"
)

func RunCodeWithTestcases(m *containers.ContainersPoolManger, code string, testcases []string, codeLanguage string, resourceLimit models.ResourceLimit) ([]string, error) {
	var outputs []string
	for i, testCase := range testcases {
		testOutput, err := RuntestCase(m, code, testCase, codeLanguage, resourceLimit)
		if err != nil {
			return nil, fmt.Errorf("testcase #%d failed: %w", i+1, err)
		}
		if testOutput != nil {
			outputs = append(outputs, *testOutput)
		}
	}
	if outputs == nil {
		outputs = []string{}
	}
	return outputs, nil
}

func RuntestCase(m *containers.ContainersPoolManger, code string, testcase string, codeLanguage string, resourceLimit models.ResourceLimit) (*string, error) {
	overallStart := time.Now()

	doc, exec, _, err := m.GetContainer(codeLanguage)
	if err != nil {
		return nil, fmt.Errorf("failed to get container: %w", err)
	}
	defer m.FreeContainer(doc)

	fileName, err := exec.CopyCodeToFile(doc, code)
	if err != nil {
		return nil, fmt.Errorf("failed to copy code: %w", err)
	}

	compileStart := time.Now()
	compileCommand, err := runStepWithTimeout(10*time.Second, func(ctx context.Context) (string, error) {
		return exec.CompileCode(doc, fileName, ctx)
	})
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %w", err)
	}
	log.Printf("Step 'Compile' completed in %v", time.Since(compileStart))

	runStart := time.Now()
	output, err := runStepWithTimeout(time.Duration(resourceLimit.TimeLimitInSeconds)*time.Second, func(ctx context.Context) (string, error) {
		return exec.RunTestCases(doc, testcase, compileCommand, ctx)
	})
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}
	log.Printf("Step 'Run' completed in %v", time.Since(runStart))

	fmt.Printf(" Total Execution Time: %v\n", time.Since(overallStart))
	return &output, nil
}

func runStepWithTimeout(timeout time.Duration, task func(ctx context.Context) (string, error)) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := task(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return "", &customErrors.TimeLimitExceededError{Limit: int(timeout.Seconds())}
		}
		return "", err
	}
	return output, nil
}
