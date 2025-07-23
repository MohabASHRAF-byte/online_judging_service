package processor

import (
	"archive/tar"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"judging-service/containers"
	"judging-service/internal/models"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
)

func RunCppWithTestcases(m *containers.ContainersPoolManger, code string, testcases []string) ([]string, error) {

	var outputs []string
	sectionStart := time.Now()
	overallStart := time.Now()
	fmt.Printf("⏱️  Section 2 (Container Create & Start): %v\n", time.Since(sectionStart))
	doc, err := m.GetContainer(models.Cpp)
	defer m.FreeContainer(doc)

	fmt.Printf(strconv.Itoa(doc.ID))
	// ============================================
	// SECTION 3: Create TAR Archive and Copy to Container
	// ============================================
	sectionStart = time.Now()

	// Create TAR archive directly from code string (no file I/O)
	tarData, err := createTarArchiveFromMemory("main.cpp", code)
	if err != nil {
		return nil, fmt.Errorf("failed to create tar archive: %v", err)
	}

	// Copy the TAR archive to the container's /workspace containers
	err = doc.Cli.CopyToContainer(doc.Ctx, doc.ContainerResp.ID, "/workspace", tarData, container.CopyToContainerOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to copy source to container: %v", err)
	}

	fmt.Printf("⏱️  Section 3 (TAR Create & Copy): %v\n", time.Since(sectionStart))

	// ============================================
	// SECTION 4: Compile C++ Code Inside Container
	// ============================================
	sectionStart = time.Now()

	// Configure compilation command execution
	compileExecConfig := container.ExecOptions{
		Cmd:          []string{"g++", "-o", "solution", "main.cpp"}, // Compile command
		AttachStdout: true,                                          // Capture stdout
		AttachStderr: true,                                          // Capture stderr
		WorkingDir:   "/workspace",                                  // Execute in workspace
	}

	// Create execution instance for compilation
	compileExecResp, err := doc.Cli.ContainerExecCreate(doc.Ctx, doc.ContainerResp.ID, compileExecConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create compile exec: %v", err)
	}

	// Start compilation and attach to streams
	compileAttachResp, err := doc.Cli.ContainerExecAttach(doc.Ctx, compileExecResp.ID, container.ExecStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to compile exec: %v", err)
	}
	defer compileAttachResp.Close()

	// Read compilation output (stdout/stderr combined)
	compileOutput, err := io.ReadAll(compileAttachResp.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read compile output: %v", err)
	}

	// Check if compilation was successful
	compileInspect, err := doc.Cli.ContainerExecInspect(doc.Ctx, compileExecResp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect compile exec: %v", err)
	}

	if compileInspect.ExitCode != 0 {
		return nil, fmt.Errorf("compilation failed: %s", string(compileOutput))
	}

	compilationTime := time.Since(sectionStart)
	fmt.Printf("⏱️  Section 4 (C++ Compilation): %v\n", compilationTime)
	fmt.Println("✓ Compilation successful")

	// ============================================
	// SECTION 5: Run Each Testcase and Collect Outputs
	// ============================================
	sectionStart = time.Now()

	for i, testInput := range testcases {
		testcaseStart := time.Now()
		fmt.Printf("Running testcase %d...\n", i+1)

		// Configure execution for running the compiled program
		runExecConfig := container.ExecOptions{
			Cmd:          []string{"./solution"}, // Run the compiled executable
			AttachStdin:  true,                   // Attach stdin to provide input
			AttachStdout: true,                   // Capture stdout
			AttachStderr: false,                  // Don't capture stderr
			WorkingDir:   "/workspace",           // Execute in workspace
		}

		// Create execution instance for running the program
		runExecResp, err := doc.Cli.ContainerExecCreate(doc.Ctx, doc.ContainerResp.ID, runExecConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create run exec for testcase %d: %v", i+1, err)
		}

		// Start execution and attach to streams
		runAttachResp, err := doc.Cli.ContainerExecAttach(doc.Ctx, runExecResp.ID, container.ExecStartOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to attach to run exec for testcase %d: %v", i+1, err)
		}

		// Send testcase input to the program via stdin
		go func() {
			defer runAttachResp.CloseWrite()
			_, err := runAttachResp.Conn.Write([]byte(testInput + "\n"))
			if err != nil {
				fmt.Printf("Warning: failed to write input for testcase %d: %v\n", i+1, err)
			}
		}()

		// Read program output with proper stream handling
		output, err := io.ReadAll(runAttachResp.Reader)
		runAttachResp.Close()

		if err != nil {
			return nil, fmt.Errorf("failed to read output for testcase %d: %v", i+1, err)
		}

		// Docker multiplexes stdout/stderr, so we need to demultiplex
		stdoutStr, stderrStr := demultiplexDockerOutput(output)

		// Use only stdout for the result
		cleanOutput := strings.TrimSpace(stdoutStr)
		outputs = append(outputs, cleanOutput)

		if stderrStr != "" {
			fmt.Printf("⚠️  Stderr for testcase %d: %s\n", i+1, stderrStr)
		}

		testcaseTime := time.Since(testcaseStart)
		fmt.Printf("✓ Testcase %d completed in %v: '%s'\n", i+1, testcaseTime, cleanOutput)
	}

	allTestcasesTime := time.Since(sectionStart)
	fmt.Printf("⏱️  Section 5 (All Testcases): %v\n", allTestcasesTime)
	cleanupStart := time.Now()
	//doc.Cli.ContainerStop(doc.Ctx, doc.ContainerResp.ID, container.StopOptions{})
	//doc.Cli.ContainerRemove(doc.Ctx, doc.ContainerResp.ID, container.RemoveOptions{Force: true})
	fmt.Printf("🧹 Container Cleanup: %v\n", time.Since(cleanupStart))
	fmt.Printf("🎯 Total Execution Time: %v\n", time.Since(overallStart))
	return outputs, nil
}

// ============================================
// HELPER FUNCTION: Create TAR Archive Directly from Memory
// ============================================
func createTarArchiveFromMemory(filename, content string) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Convert string content to bytes (no file I/O)
	fileContent := []byte(content)

	// Create TAR header for the file
	header := &tar.Header{
		Name: filename,                // File name in the archive
		Mode: 0644,                    // File permissions
		Size: int64(len(fileContent)), // File size
	}

	// Write header to TAR archive
	if err := tw.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("failed to write tar header: %v", err)
	}

	// Write file content to TAR archive (directly from memory)
	if _, err := tw.Write(fileContent); err != nil {
		return nil, fmt.Errorf("failed to write file content to tar: %v", err)
	}

	// Close TAR writer
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %v", err)
	}

	return &buf, nil
}

func demultiplexDockerOutput(data []byte) (stdout, stderr string) {
	var stdoutBuf, stderrBuf bytes.Buffer

	for len(data) > 8 {
		// Docker uses 8-byte headers: [stream_type][0][0][0][size_bytes]
		streamType := data[0]
		size := binary.BigEndian.Uint32(data[4:8])

		if len(data) < 8+int(size) {
			break
		}

		payload := data[8 : 8+size]

		switch streamType {
		case 1: // stdout
			stdoutBuf.Write(payload)
		case 2: // stderr
			stderrBuf.Write(payload)
		}

		data = data[8+size:]
	}

	return stdoutBuf.String(), stderrBuf.String()
}
