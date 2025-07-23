package service

import (
	"fmt"
	"github.com/docker/docker/api/types/container"
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
