package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"io"
	"judging-service/internal/customErrors"
	"judging-service/internal/models"
	"log"
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

// ctxReader wraps an io.Reader to make it respect a context's deadline.
func ctxReader(ctx context.Context, r io.Reader) io.Reader {
	return &cancellableReader{ctx: ctx, r: r}
}

type cancellableReader struct {
	ctx context.Context
	r   io.Reader
}

func (cr *cancellableReader) Read(p []byte) (int, error) {
	// Check if context is already done.
	if err := cr.ctx.Err(); err != nil {
		return 0, err
	}
	// Perform the read in a goroutine so we can select on the context.
	type result struct {
		n   int
		err error
	}
	done := make(chan result, 1)
	go func() {
		n, err := cr.r.Read(p)
		done <- result{n, err}
	}()

	select {
	case res := <-done:
		return res.n, res.err
	case <-cr.ctx.Done():
		return 0, cr.ctx.Err()
	}
}

func (_ CppRunLangInterFace) CompileCode(containerCpy *models.Container, fileName string, ctx context.Context) (string, error) {
	log.Println("[DEBUG: 1] Entering CompileCode function.")

	var executableFileCommand = "./solution"
	compileExecConfig := container.ExecOptions{
		Cmd:          []string{"g++", "-o", "solution", fileName},
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/workspace",
	}
	log.Println("[DEBUG: 2] Compile exec config created.")

	log.Println("[DEBUG: 3] Attempting to create compile exec in container...")
	compileExecResp, err := containerCpy.Cli.ContainerExecCreate(ctx, containerCpy.ContainerResp.ID, compileExecConfig)

	// NOTE: This error handling logic is preserved from your original code.
	// It is un-idiomatic but logged for clarity.
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		log.Printf("[DEBUG: 3.1] Caught timeout error during ContainerExecCreate: %v", err)
		return "", &customErrors.TimeLimitExceededError{Limit: 1}
	}

	if err != nil {
		log.Printf("[DEBUG: 3.2] Caught a non-timeout error during ContainerExecCreate: %v", err)
		return "", fmt.Errorf("failed to create compile exec: %v", err)
	}
	log.Printf("[DEBUG: 4] Successfully created compile exec with ID: %s", compileExecResp.ID)

	log.Println("[DEBUG: 5] Attaching to compile exec streams...")
	compileAttachResp, err := containerCpy.Cli.ContainerExecAttach(ctx, compileExecResp.ID, container.ExecStartOptions{})
	if err != nil {
		log.Printf("[DEBUG: 5.1] Error attaching to compile exec: %v", err)
		return "", fmt.Errorf("failed to attach to compile exec: %v", err)
	}
	defer compileAttachResp.Close()
	log.Println("[DEBUG: 6] Successfully attached to compile exec.")

	log.Println("[DEBUG: 7] Reading all output from compile exec... (This is where it will likely hang if the context expires)")
	compileOutputBytes, err := io.ReadAll(compileAttachResp.Reader)

	if errors.Is(err, context.DeadlineExceeded) {
		// THIS BLOCK WILL LIKELY NEVER BE REACHED because io.ReadAll doesn't respect the context.
		log.Printf("[DEBUG: 7.1] Caught context deadline exceeded error after io.ReadAll returned.")
		return "", err
	} else if err != nil {
		log.Printf("[DEBUG: 7.2] Caught a non-timeout error during io.ReadAll: %v", err)
		return "", fmt.Errorf("failed to read compile output: %v", err)
	}
	log.Printf("[DEBUG: 8] Finished reading from exec stream. Read %d bytes.", len(compileOutputBytes))

	log.Println("[DEBUG: 9] Inspecting exec result...")
	compileInspect, err := containerCpy.Cli.ContainerExecInspect(ctx, compileExecResp.ID)
	if err != nil {
		log.Printf("[DEBUG: 9.1] Error inspecting compile exec: %v", err)
		return "", fmt.Errorf("failed to inspect compile exec: %v", err)
	}
	log.Printf("[DEBUG: 10] Exec inspect complete. ExitCode: %d", compileInspect.ExitCode)

	if compileInspect.ExitCode != 0 {
		log.Printf("[DEBUG: 10.1] Compilation failed with non-zero exit code. Output: %s", string(compileOutputBytes))
		return "", &customErrors.CompilationError{}
	}
	log.Println("[DEBUG: 11] Compilation successful. Returning executable command.")

	return executableFileCommand, nil
}

func (_ CppRunLangInterFace) RunTestCases(containerCpy *models.Container, testcase string, compileCommand string, ctx context.Context) (string, error) {
	log.Println("[DEBUG:12] ")

	testcaseStart := time.Now()
	runExecConfig := container.ExecOptions{
		Cmd:          []string{compileCommand},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: false,
		WorkingDir:   "/workspace",
	}
	log.Println("[DEBUG:13] ")

	runExecResp, err := containerCpy.Cli.ContainerExecCreate(ctx, containerCpy.ContainerResp.ID, runExecConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create run exec for testcase : %v", err)
	}
	log.Println("[DEBUG:14] ")

	runAttachResp, err := containerCpy.Cli.ContainerExecAttach(ctx, runExecResp.ID, container.ExecStartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to attach to run exec for testcase : %v", err)
	}
	defer runAttachResp.Close()
	log.Println("[DEBUG:15] ")

	go func() {
		defer runAttachResp.CloseWrite()
		_, err := runAttachResp.Conn.Write([]byte(testcase + "\n"))
		if err != nil {
			fmt.Printf("Warning: failed to write input for testcase: %v\n", err)
		}
	}()
	log.Println("[DEBUG:16] ")

	output, err := io.ReadAll(ctxReader(ctx, runAttachResp.Reader))
	log.Println("[DEBUG:17] ")

	if err != nil {
		return "", err
	}

	stdoutStr, stderrStr := demultiplexDockerOutput(output)

	cleanOutput := strings.TrimSpace(stdoutStr)
	log.Println("[DEBUG:18] ")

	if stderrStr != "" {
		fmt.Printf("⚠️  Stderr for testcase: %s\n", stderrStr)
	}

	testcaseTime := time.Since(testcaseStart)
	fmt.Printf("✓ Testcase completed in: %s. Output: '%s'\n", testcaseTime, cleanOutput)

	return cleanOutput, nil
}
