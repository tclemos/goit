package goit

import (
	"context"
	"fmt"
	"strings"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// We need these public variables to share information betwee
// TestMain and OtherTests, if you have a better idea, tell me
var (
	Ctx context.Context
)

// Container represents a docker container
type Container interface {

	// Executed after the container is started, use it to run migrations
	// copy files, etc
	AfterStart(context.Context, *dockertest.Resource) error
}

// ContainerFromRepository represents a docker container that will be
// created based on an docker image from a docker repository
type containerFromRepository interface {
	Container

	// Options to execute the container
	Options() (*dockertest.RunOptions, error)
}

// ContainerFromDockerFile represents a docker container that will be
// created based on a dockerfile
type containerFromDockerFile interface {
	Container

	// The Name of the container
	ContainerName() string

	// Path of the dockerfile on disk
	DockerFilePath() string

	// Arguments used during the build phase
	BuildArgs() []docker.BuildArg

	// Port Bindings for the container
	PortBindings() map[docker.Port][]docker.PortBinding
}

type ContainerParams struct {
	Repository string
	Tag        string
	Env        []string
}

func (p ContainerParams) GetRepoTag(defaultRepo, defaultTag string) (repo, tag string) {
	if strings.TrimSpace(p.Repository) != "" {
		repo = p.Repository
	} else {
		repo = defaultRepo
	}

	if strings.TrimSpace(p.Tag) != "" {
		tag = p.Tag
	} else {
		tag = defaultTag
	}

	return repo, tag
}

func (p ContainerParams) MergeEnv(env []string) []string {
	envMap := map[string]string{}
	getEnvKeyValue := func(s string) (string, string) {
		parts := strings.Split(s, "=")
		return parts[0], parts[1]
	}

	for _, e := range env {
		k, v := getEnvKeyValue(e)
		envMap[k] = v
	}

	for _, e := range p.Env {
		k, v := getEnvKeyValue(e)
		envMap[k] = v
	}

	newEnv := make([]string, len(envMap))

	i := 0
	for k, v := range envMap {
		newEnv[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}

	return newEnv
}
