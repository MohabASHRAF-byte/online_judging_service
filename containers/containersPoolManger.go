package containers

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"judging-service/internal/models"
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

func (manger *ContainersPoolManger) getContainer(language models.Language) (*models.Container, error) {
	manger.mu.Lock()
	defer manger.mu.Unlock()

	sort.Slice(manger.FreeContainers, func(i, j int) bool {
		return manger.FreeContainers[i].LastModified.Before(manger.FreeContainers[j].LastModified)
	})

	for i, c := range manger.FreeContainers {
		if c.IsEmpty && c.Language == language {
			c.LastModified = time.Now()
			c.IsEmpty = false
			manger.FreeContainers[i] = c
			return c, nil
		}
	}

	if len(manger.FreeContainers) < manger.Limit {
		newContainer, err := manger.newDockerContainer(language)
		if err != nil {
			return nil, err
		}
		manger.FreeContainers = append(manger.FreeContainers, newContainer)
		return newContainer, nil
	} else {
		for i, c := range manger.FreeContainers {
			if c.IsEmpty {
				manger.FreeContainers[i] = manger.FreeContainers[len(manger.FreeContainers)-1]
				manger.FreeContainers = manger.FreeContainers[:len(manger.FreeContainers)-1]
				newContainer, err := manger.newDockerContainer(language)
				if err != nil {
					return nil, err
				}
				manger.FreeContainers = append(manger.FreeContainers, newContainer)
				return newContainer, nil
			}
		}
	}

	return nil, fmt.Errorf("all containers are busy and pool is full")
}

func (m *ContainersPoolManger) GetContainer(language models.Language) (*models.Container, error) {
	maxAttempts := 50
	sleepDuration := 1000 * time.Millisecond
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		doc, err := m.getContainer(language)
		if err == nil {
			return doc, nil
		}
		lastErr = err
		time.Sleep(sleepDuration)
	}
	return nil, fmt.Errorf("after %d attempts, last error: %w", maxAttempts, lastErr)
}

func (manger *ContainersPoolManger) FreeContainer(container *models.Container) {
	manger.mu.Lock()
	defer manger.mu.Unlock()

	for i, c := range manger.FreeContainers {
		if c.ID == container.ID {
			manger.FreeContainers[i].IsEmpty = true
			return
		}
	}
}

func (manger *ContainersPoolManger) newDockerContainer(lang models.Language) (*models.Container, error) {

	var docker = &models.Container{
		Ctx:          context.Background(),
		ID:           manger.NextID,
		Language:     lang,
		IsEmpty:      false,
		IsInit:       true,
		LastModified: time.Now(),
	}
	manger.NextID++
	//
	var dockerImage models.LanguageDockerImageName

	if lang == models.Cpp {
		dockerImage = models.CppImage
	} else if lang == models.Python {
		dockerImage = models.PythonImage
	}
	//
	var err error
	docker.Cli, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}
	docker.Cli.NegotiateAPIVersion(docker.Ctx)
	containerConfig := &container.Config{
		Image:      string(dockerImage),
		Tty:        false,
		Cmd:        []string{"sleep", "600"},
		WorkingDir: "/workspace",
	}
	hostConfig := &container.HostConfig{}
	docker.ContainerResp, err = docker.Cli.ContainerCreate(docker.Ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}
	err = docker.Cli.ContainerStart(docker.Ctx, docker.ContainerResp.ID, container.StartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}
	return docker, nil
}
