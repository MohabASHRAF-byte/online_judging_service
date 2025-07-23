package service

import (
	"fmt"
	"github.com/docker/docker/api/types/container"
	"io"
	"judging-service/internal/models"
	"strings"
	"time"
)

type CppRunLangInterFace struct {
}

// the target of this function is only to pass the name of the file and call the util file to copy
func (_ CppRunLangInterFace) CopyCodeToFile(containerCpy *models.Container, code string) (string, error) {
	_, err := CopyCodeToFileGlobalUtil(containerCpy, "main.cpp", code)
	if err != nil {
		return "", err
	}
	return "main.cpp", nil

}

func (_ CppRunLangInterFace) CompileCode(containerCpy *models.Container, fileName string) (string, error) {

	var executableFileCommand = "./solution"
	compileExecConfig := container.ExecOptions{
		Cmd:          []string{"g++", "-o", "solution", fileName},
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/workspace",
	}

	compileExecResp, err := containerCpy.Cli.ContainerExecCreate(containerCpy.Ctx, containerCpy.ContainerResp.ID, compileExecConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create compile exec: %v", err)
	}

	compileAttachResp, err := containerCpy.Cli.ContainerExecAttach(containerCpy.Ctx, compileExecResp.ID, container.ExecStartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to attach to compile exec: %v", err)
	}
	defer compileAttachResp.Close()

	compileOutput, err := io.ReadAll(compileAttachResp.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to read compile output: %v", err)
	}

	compileInspect, err := containerCpy.Cli.ContainerExecInspect(containerCpy.Ctx, compileExecResp.ID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect compile exec: %v", err)
	}

	if compileInspect.ExitCode != 0 {
		return "", fmt.Errorf("compilation failed: %s", string(compileOutput))
	}

	return executableFileCommand, nil
}

func (_ CppRunLangInterFace) RunTestCases(containerCpy *models.Container, testcase string, compileCommand string) (string, error) {

	//var output string

	testcaseStart := time.Now()
	runExecConfig := container.ExecOptions{
		Cmd:          []string{compileCommand}, // Run the compiled executable
		AttachStdin:  true,                     // Attach stdin to provide input
		AttachStdout: true,                     // Capture stdout
		AttachStderr: false,                    // Don't capture stderr
		WorkingDir:   "/workspace",             // Execute in workspace
	}

	// Create execution instance for running the program
	runExecResp, err := containerCpy.Cli.ContainerExecCreate(containerCpy.Ctx, containerCpy.ContainerResp.ID, runExecConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create run exec for testcase : %v", err)
	}

	// Start execution and attach to streams
	runAttachResp, err := containerCpy.Cli.ContainerExecAttach(containerCpy.Ctx, runExecResp.ID, container.ExecStartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to attach to run exec for testcase : %v", err)
	}

	// Send testcase input to the program via stdin
	go func() {
		defer runAttachResp.CloseWrite()
		_, err := runAttachResp.Conn.Write([]byte(testcase + "\n"))
		if err != nil {
			fmt.Printf("Warning: failed to write input for testcase: %v\n", err)
		}
	}()

	// Read program output with proper stream handling
	output, err := io.ReadAll(runAttachResp.Reader)
	runAttachResp.Close()

	if err != nil {
		return "", fmt.Errorf("failed to read output for testcase %v", err)
	}

	// Docker multiplexes stdout/stderr, so we need to demultiplex
	stdoutStr, stderrStr := demultiplexDockerOutput(output)

	// Use only stdout for the result
	cleanOutput := strings.TrimSpace(stdoutStr)

	if stderrStr != "" {
		fmt.Printf("⚠️  Stderr for testcase: %s\n", stderrStr)
	}

	testcaseTime := time.Since(testcaseStart)
	fmt.Printf("✓ Testcase %d completed in: '%s'\n", testcaseTime, cleanOutput)

	return cleanOutput, nil
}
