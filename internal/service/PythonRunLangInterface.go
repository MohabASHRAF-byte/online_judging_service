package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"io"
	"judging-service/internal/customErrors"
	"judging-service/internal/models"
	"strings"
	"time"
)

type PythonRunLangInterface struct {
}

func (_ PythonRunLangInterface) CopyCodeToFile(containerCpy *models.Container, code string) (string, error) {
	_, err := CopyCodeToFileGlobalUtil(containerCpy, "main.py", code)
	if err != nil {
		return "", err
	}
	return "main.py", nil
}

func (_ PythonRunLangInterface) CompileCode(containerCpy *models.Container, fileName string, ctx context.Context) (string, error) {

	var executableFileCommand = "python " + fileName
	compileExecConfig := container.ExecOptions{
		Cmd:          []string{"python", "-m", "py_compile", fileName},
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/workspace",
	}
	compileExecResp, err := containerCpy.Cli.ContainerExecCreate(ctx, containerCpy.ContainerResp.ID, compileExecConfig)
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		return "", &customErrors.TimeLimitExceededError{Limit: 1}
	}
	if err != nil {
		return "", fmt.Errorf("failed to create compile exec: %v", err)
	}
	compileAttachResp, err := containerCpy.Cli.ContainerExecAttach(ctx, compileExecResp.ID, container.ExecStartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to attach to compile exec: %v", err)
	}
	defer compileAttachResp.Close()
	_, err = io.ReadAll(compileAttachResp.Reader)
	if errors.Is(err, context.DeadlineExceeded) {
		return "", err
	} else if err != nil {
		return "", fmt.Errorf("failed to read compile output: %v", err)
	}
	compileInspect, err := containerCpy.Cli.ContainerExecInspect(ctx, compileExecResp.ID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect compile exec: %v", err)
	}
	if compileInspect.ExitCode != 0 {
		return "", &customErrors.CompilationError{}
	}
	return executableFileCommand, nil
}

func (_ PythonRunLangInterface) RunTestCases(containerCpy *models.Container, testcase string, compileCommand string, ctx context.Context) (string, error) {

	testcaseStart := time.Now()
	cmdParts := strings.Fields(compileCommand)

	runExecConfig := container.ExecOptions{
		Cmd:          cmdParts,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: false,
		WorkingDir:   "/workspace",
	}

	runExecResp, err := containerCpy.Cli.ContainerExecCreate(ctx, containerCpy.ContainerResp.ID, runExecConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create run exec for testcase : %v", err)
	}

	runAttachResp, err := containerCpy.Cli.ContainerExecAttach(ctx, runExecResp.ID, container.ExecStartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to attach to run exec for testcase : %v", err)
	}
	defer runAttachResp.Close()

	go func() {
		defer runAttachResp.CloseWrite()
		_, err := runAttachResp.Conn.Write([]byte(testcase + "\n"))
		if err != nil {
			fmt.Printf("Warning: failed to write input for testcase: %v\n", err)
		}
	}()
	output, err := io.ReadAll(ctxReader(ctx, runAttachResp.Reader))
	if err != nil {
		return "", err
	}

	stdoutStr, stderrStr := demultiplexDockerOutput(output)

	cleanOutput := strings.TrimSpace(stdoutStr)

	if stderrStr != "" {
		fmt.Printf("Stderr for testcase: %s\n", stderrStr)
	}
	testcaseTime := time.Since(testcaseStart)
	fmt.Printf("Testcase completed in: %s. Output: '%s'\n", testcaseTime, cleanOutput)

	return cleanOutput, nil
}
