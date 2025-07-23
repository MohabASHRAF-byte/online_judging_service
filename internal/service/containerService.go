package service

import (
	"archive/tar"
	"bytes"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"io"
	"judging-service/internal/models"
)

type CppRunLangInterFace struct {
}

func (_ CppRunLangInterFace) CopyCodeToFile(containerCpy *models.Container, language models.Language, code string) (string, error) {
	// Create TAR archive directly from code string (no file I/O)
	var fileName = ""
	if language == models.Cpp {
		fileName = "main.cpp"
	} else if language == models.Python {
		fileName = "main.py"
	}
	tarData, err := createTarArchiveFromMemory(fileName, code)
	if err != nil {
		return "", fmt.Errorf("failed to create tar archive: %v", err)
	}

	// Copy the TAR archive to the container's /workspace containers
	err = containerCpy.Cli.CopyToContainer(containerCpy.Ctx, containerCpy.ContainerResp.ID, "/workspace", tarData, container.CopyToContainerOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to copy source to container: %v", err)

	}
	return fileName, nil
}

func (_ CppRunLangInterFace) CompileCode() {

}

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
