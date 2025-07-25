package processor

import (
	"context"
	"errors"
	"fmt"
	"judging-service/api/Dtos"
	"judging-service/containers"
	customErrors "judging-service/internal/customErrors"
	"judging-service/internal/models"
	"log"
	"strings"
	"time"
)

func RunCodeWithTestcases(
	m *containers.ContainersPoolManger,
	submission Dtos.SubmissionQueueDto,
) (models.JudgingResult, error) {

	outputs := make([]models.TestCaseOutput, 0, len(submission.InputTests))
	for i, testCase := range submission.InputTests {
		testOutput, err := RuntestCase(m, submission.Code, testCase.Input, submission.Language, models.ResourceLimit{
			MemoryLimitInMB:    submission.MemoryLimit,
			TimeLimitInSeconds: submission.TimeLimit,
			CPU:                1,
		})
		if err != nil {
			var verdict int = 3

			if strings.Contains(err.Error(), "Time Limit Exceeded") {
				verdict = 2
			}
			return models.JudgingResult{
				SubmissionId: submission.SubmissionId,
				Verdict:      verdict,
				Outputs:      nil,
				IsErrorExist: true,
				FallingTest:  i + 1,
			}, fmt.Errorf("testcase #%d failed: %w", i+1, err)
		}

		if testOutput != nil {
			outputs = append(outputs, models.TestCaseOutput{
				Output:     *testOutput,
				TestCaseId: testCase.TestCaseId,
			})
		}
	}
	return models.JudgingResult{
		SubmissionId: submission.SubmissionId,
		Verdict:      0,
		IsErrorExist: false,
		Outputs:      outputs,
		FallingTest:  0,
	}, nil
}
func RuntestCase(m *containers.ContainersPoolManger, code string, testcase string, codeLanguage int, resourceLimit models.ResourceLimit) (*string, error) {
	overallStart := time.Now()

	doc, exec, _, err := m.GetContainerWithLimits(codeLanguage, resourceLimit)
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
