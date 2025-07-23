package service

import (
	"archive/tar"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"io"
	"judging-service/internal/models"
)

func CopyCodeToFileGlobalUtil(containerCpy *models.Container, fileName string, code string) (string, error) {
	tarData, err := createTarArchiveFromMemory(fileName, code)
	if err != nil {
		return "", fmt.Errorf("failed to create tar archive: %v", err)
	}

	err = containerCpy.Cli.CopyToContainer(containerCpy.Ctx, containerCpy.ContainerResp.ID, "/workspace", tarData, container.CopyToContainerOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to copy source to container: %v", err)

	}
	return fileName, nil
}

func createTarArchiveFromMemory(filename, content string) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	fileContent := []byte(content)

	header := &tar.Header{
		Name: filename,
		Mode: 0644,
		Size: int64(len(fileContent)),
	}

	if err := tw.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("failed to write tar header: %v", err)
	}

	if _, err := tw.Write(fileContent); err != nil {
		return nil, fmt.Errorf("failed to write file content to tar: %v", err)
	}

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
