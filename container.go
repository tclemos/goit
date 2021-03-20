package goit

import (
	"context"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/tclemos/goit/log"
)

// We need these public variables to share information betwee
// TestMain and OtherTests, if you have a better idea, tell me
var (
	Ctx context.Context
)

// Container represents a docker container
type Container interface {
	// Options to execute the container
	Options() (*dockertest.RunOptions, error)

	// Executed after the container is started, use it to run migrations
	// copy files, etc
	AfterStart(context.Context, *dockertest.Resource) error
}

// startContainer creates and initializes a container accordingly to the provided options
func startContainer(ctx context.Context, p *dockertest.Pool, o *dockertest.RunOptions) (*dockertest.Resource, error) {
	log.Logf("starting container: %s", o.Name)
	r, err := p.RunWithOptions(o, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Error(err, "failed to start container, check if docker is running and exposing deamon on tcp://localhost:2375")
		return nil, err
	}

	err = r.Expire(60)
	if err != nil {
		log.Errorf(err, "could not setup container to expire: %s", o.Name)
		return nil, err
	}

	log.Logf("container started: %s", o.Name)
	return r, nil
}
