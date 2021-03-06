package dockerfile

import (
	"context"
	"fmt"
	"strings"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// Params needed to start a container from a dockerfile
type Params struct {
	ContainerName  string
	DockerFilePath string
	Env            map[string]string
	BuildArgs      []docker.BuildArg
	PortBindings   map[docker.Port][]docker.PortBinding
	AfterStart     func(context.Context, *dockertest.Resource, *map[string]interface{}) error
}

// Container metadata to load a container
type Container struct {
	params Params
	Values map[string]interface{}
}

// NewContainer creates a new instance of Container
func NewContainer(p Params) *Container {
	if strings.TrimSpace(p.ContainerName) == "" {
		panic("ContainerName is required")
	}

	return &Container{
		params: p,
	}
}

func (c *Container) ContainerName() string {
	return c.params.ContainerName
}

func (c *Container) DockerFilePath() string {
	return c.params.DockerFilePath
}

func (c *Container) Env() []string {
	env := make([]string, 0, len(c.params.Env))

	for k, v := range c.params.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}

func (c *Container) BuildArgs() []docker.BuildArg {
	return c.params.BuildArgs
}

func (c *Container) PortBindings() map[docker.Port][]docker.PortBinding {
	return c.params.PortBindings
}

// AfterStart will check the connection and execute migrations
func (c *Container) AfterStart(ctx context.Context, r *dockertest.Resource) error {
	if c.params.AfterStart != nil {
		return c.params.AfterStart(ctx, r, &c.Values)
	}
	return nil
}
