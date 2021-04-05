package goit

import (
	"context"
	"fmt"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/tclemos/goit/log"
)

var (
	pool      *dockertest.Pool
	resources []*dockertest.Resource
)

type Options struct {
	ExpireContainersAfterSeconds uint
}

func Start(ctx context.Context, containers ...Container) {
	StartWithOptions(ctx, Options{
		ExpireContainersAfterSeconds: 60,
	}, containers...)
}

// Start the integration test environment
func StartWithOptions(ctx context.Context, opt Options, containers ...Container) {
	log.Log("initializing containers")
	resources = []*dockertest.Resource{}

	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Errorf(err, "failed to create docker pool")
		panic(err)
	}

	for _, c := range containers {
		o, err := c.Options()
		log.Logf("loading container with options: %v", o)
		handleContainerErr(err, "can't load container")

		r, err := startContainer(ctx, pool, o, opt)
		handleContainerErr(err, "can't start container")

		log.Logf("executing AfterStart for container: %s", r.Container.Name)
		err = c.AfterStart(ctx, r)
		handleContainerErr(err, fmt.Sprintf("failed to execute AfterStarted for container: %s", r.Container.Name))

		resources = append(resources, r)
	}
}

// Stop the integration test environment
func Stop() {
	for _, r := range resources {
		log.Logf("purging container: %s", r.Container.Name)
		err := pool.Purge(r)
		if err != nil {
			log.Errorf(err, "could not purge container: %v", r.Container.Name)
		} else {
			log.Logf("container purged: %s", r.Container.Name)
		}
	}
}

// startContainer creates and initializes a container accordingly to the provided options
func startContainer(ctx context.Context, p *dockertest.Pool, ropt *dockertest.RunOptions, opt Options) (*dockertest.Resource, error) {
	log.Logf("starting container: %s", ropt.Name)
	r, err := p.RunWithOptions(ropt, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Error(err, "failed to start container, check if docker is running and exposing deamon on tcp://localhost:2375")
		return nil, err
	}

	err = r.Expire(opt.ExpireContainersAfterSeconds)
	if err != nil {
		log.Errorf(err, "could not setup container to expire: %s", ropt.Name)
		return nil, err
	}

	log.Logf("container started: %s", ropt.Name)
	return r, nil
}

func handleContainerErr(err error, m string, args ...interface{}) {
	if err != nil {
		log.Errorf(err, m, args...)
		Stop()
		panic(err)
	}
}
