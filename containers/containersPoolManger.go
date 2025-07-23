package containers

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"judging-service/internal/models"
	"judging-service/internal/service"
)

type ContainersPoolManger struct {
	Limit          int
	FreeContainers []*models.Container
	NextID         int
	mu             sync.Mutex
}

func NewContainersPoolManger(limit int) *ContainersPoolManger {
	return &ContainersPoolManger{
		Limit:  limit,
		NextID: 1,
	}
}

// GetContainerWithLimits is the new primary method for acquiring a container.
// It handles language selection and retries, calling the internal get-or-create logic.
func (m *ContainersPoolManger) GetContainerWithLimits(language string, limit models.ResourceLimit) (*models.Container, models.LangContainer, models.Language, error) {
	maxAttempts := 50
	sleepDuration := 1000 * time.Millisecond

	var exec models.LangContainer
	var lang models.Language
	switch language {
	case string(models.Cpp):
		lang = models.Cpp
		exec = service.CppRunLangInterFace{}
	case string(models.Python):
		lang = models.Python
		exec = service.PythonRunLangInterface{}

	// Add other languages here
	default:
		return nil, nil, "", fmt.Errorf("invalid language: %q", language)
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		doc, err := m.getOrCreateContainer(lang, limit)
		if err == nil {
			return doc, exec, lang, nil
		}
		lastErr = err
		time.Sleep(sleepDuration)
	}
	return nil, nil, "", fmt.Errorf("after %d attempts, last error: %w", maxAttempts, lastErr)
}

// getOrCreateContainer implements the logic to always create a new container,
// evicting an old one if the pool is full.
func (manger *ContainersPoolManger) getOrCreateContainer(language models.Language, limit models.ResourceLimit) (*models.Container, error) {
	manger.mu.Lock()
	defer manger.mu.Unlock()

	// If the pool is full, find and evict an old, empty container.
	if len(manger.FreeContainers) >= manger.Limit {
		log.Println("Pool is full. Searching for a container to evict.")
		// Sort to find the oldest container reliably.
		sort.Slice(manger.FreeContainers, func(i, j int) bool {
			return manger.FreeContainers[i].LastModified.Before(manger.FreeContainers[j].LastModified)
		})

		evictedIndex := -1
		for i, c := range manger.FreeContainers {
			if c.IsEmpty {
				evictedIndex = i
				break
			}
		}

		if evictedIndex == -1 {
			return nil, fmt.Errorf("all containers are busy and pool is full")
		}

		oldContainer := manger.FreeContainers[evictedIndex]
		// Efficiently remove the container from the slice.
		manger.FreeContainers[evictedIndex] = manger.FreeContainers[len(manger.FreeContainers)-1]
		manger.FreeContainers = manger.FreeContainers[:len(manger.FreeContainers)-1]

		// Asynchronously clean up the old container.
		go manger.removeContainer(oldContainer)
		log.Printf("Evicted container ID %d. Creating new container.", oldContainer.ID)
	}

	// Create and add the new container.
	return manger.createAndAddContainer(language, limit)
}

// createAndAddContainer is a helper to create and append a new container.
func (manger *ContainersPoolManger) createAndAddContainer(lang models.Language, limit models.ResourceLimit) (*models.Container, error) {
	newContainer, err := manger.newDockerContainer(lang, limit)
	if err != nil {
		return nil, err
	}
	manger.FreeContainers = append(manger.FreeContainers, newContainer)
	return newContainer, nil
}

// removeContainer stops and removes a Docker container in the background.
func (manger *ContainersPoolManger) removeContainer(c *models.Container) {
	log.Printf("Scheduling asynchronous removal of container ID: %s", c.ContainerResp.ID)
	ctx := context.Background()
	_ = c.Cli.ContainerStop(ctx, c.ContainerResp.ID, container.StopOptions{})
	_ = c.Cli.ContainerRemove(ctx, c.ContainerResp.ID, container.RemoveOptions{Force: true})
	log.Printf("Asynchronous removal of container ID %s completed.", c.ContainerResp.ID)
}

func (manger *ContainersPoolManger) FreeContainer(container *models.Container) {
	manger.mu.Lock()
	defer manger.mu.Unlock()

	for i, c := range manger.FreeContainers {
		if c.ID == container.ID {
			manger.FreeContainers[i].IsEmpty = true
			manger.FreeContainers[i].LastModified = time.Now()
			return
		}
	}
}

// newDockerContainer now accepts resource limits.
func (manger *ContainersPoolManger) newDockerContainer(lang models.Language, limit models.ResourceLimit) (*models.Container, error) {
	docker := &models.Container{
		Ctx:          context.Background(),
		ID:           manger.NextID,
		Language:     lang,
		IsEmpty:      false,
		IsInit:       true,
		LastModified: time.Now(),
	}
	manger.NextID++

	var dockerImage models.LanguageDockerImageName
	switch lang {
	case models.Cpp:
		dockerImage = models.CppImage
	case models.Python:
		dockerImage = models.PythonImage
	default:
		return nil, fmt.Errorf("unsupported language for docker image: %s", lang)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}
	docker.Cli = cli
	docker.Cli.NegotiateAPIVersion(docker.Ctx)

	containerConfig := &container.Config{
		Image:      string(dockerImage),
		Tty:        false,
		Cmd:        []string{"sleep", "600"},
		WorkingDir: "/workspace",
	}

	// Apply resource limits to the HostConfig.
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:   int64(limit.MemoryLimitInMB) * 1024 * 1024,
			CPUCount: int64(limit.CPU),
		},
	}

	resp, err := docker.Cli.ContainerCreate(docker.Ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}
	docker.ContainerResp = resp

	err = docker.Cli.ContainerStart(docker.Ctx, docker.ContainerResp.ID, container.StartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}
	return docker, nil
}
