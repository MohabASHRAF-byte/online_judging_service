package models

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"time"
)

type Container struct {
	ID            int                      `json:"id"`
	Language      Language                 `json:"language"`
	IsEmpty       bool                     `json:"is_empty"`
	IsInit        bool                     `json:"is_init"`
	LastModified  time.Time                `json:"last_modified"`
	Ctx           context.Context          `json:"-"`
	Cli           *client.Client           `json:"-"`
	ContainerResp container.CreateResponse `json:"-"`
}
