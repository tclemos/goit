package goit

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/tclemos/goit/log"
)

var (
	pool      *dockertest.Pool
	resources []*dockertest.Resource
)

type Options struct {
	// AutoRemoveContainers set the containers to remove itself when finished, e.g. docker run --rm
	AutoRemoveContainers bool

	// RestartContainers define if a container must restart after it is finished
	RestartContainers bool

	// ExpireContainersAfterSeconds sets a container to be destroid after an amount of seconds
	ExpireContainersAfterSeconds uint
}

func Start(ctx context.Context, containers ...Container) {
	StartWithOptions(ctx, DefaultOptions(), containers...)
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
		var r *dockertest.Resource
		switch cf := c.(type) {
		case containerFromDockerFile:
			r, err = startContainerFromDockerFile(ctx, pool, cf, opt)
			handleContainerErr(err, "can't start container")
		case containerFromRepository:
			r, err = startContainerFromRepository(ctx, pool, cf, opt)
			handleContainerErr(err, "can't start container")
		default:
			panic("unknown container type")
		}

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

func Run(m *testing.M) int {
	defer func() {
		err := recover()
		if recover() != nil {
			Stop()
		}
		if err != nil {
			panic(fmt.Sprintf("rethrowing panic after Stoping containers, err: %v", err))
		}
	}()
	return m.Run()
}

func DefaultOptions() Options {
	return Options{
		AutoRemoveContainers:         true,
		RestartContainers:            false,
		ExpireContainersAfterSeconds: 60,
	}
}

// startContainer creates and initializes a container accordingly to the provided options
func startContainerFromDockerFile(ctx context.Context, p *dockertest.Pool, c containerFromDockerFile, opt Options) (*dockertest.Resource, error) {
	log.Logf("starting new container")
	dir, file := filepath.Split(c.DockerFilePath())
	r, err := p.BuildAndRunWithBuildOptions(&dockertest.BuildOptions{
		ContextDir: dir,
		Dockerfile: file,
		BuildArgs:  c.BuildArgs(),
	}, &dockertest.RunOptions{
		Name:         c.ContainerName(),
		Env:          c.Env(),
		PortBindings: c.PortBindings(),
	}, getHostConfig(opt))
	if err != nil {
		log.Error(err, "failed to start container, check if docker is running and exposing deamon on tcp://localhost:2375")
		return nil, err
	}

	err = r.Expire(opt.ExpireContainersAfterSeconds)
	if err != nil {
		log.Errorf(err, "could not setup container to expire: %s", r.Container.Name)
		return nil, err
	}

	log.Logf("container started: %s", r.Container.Name)
	return r, nil
}

// startContainerFromRepository creates and initializes a container accordingly to the provided options
func startContainerFromRepository(ctx context.Context, p *dockertest.Pool, c containerFromRepository, opt Options) (*dockertest.Resource, error) {

	log.Logf("starting new container")

	o, err := c.Options()
	log.Logf("loading container with options: %v", o)
	handleContainerErr(err, "can't load container")

	r, err := p.RunWithOptions(o, getHostConfig(opt))
	if err != nil {
		log.Error(err, "failed to start container, check if docker is running and exposing deamon on tcp://localhost:2375")
		return nil, err
	}

	err = r.Expire(opt.ExpireContainersAfterSeconds)
	if err != nil {
		log.Errorf(err, "could not setup container to expire: %s", r.Container.Name)
		return nil, err
	}

	log.Logf("container started: %s", r.Container.Name)
	return r, nil
}

func getHostConfig(opt Options) func(*docker.HostConfig) {
	var restartPolicyName string
	if opt.RestartContainers {
		restartPolicyName = "yes"
	} else {
		restartPolicyName = "no"
	}

	return func(config *docker.HostConfig) {
		config.AutoRemove = opt.AutoRemoveContainers
		config.RestartPolicy = docker.RestartPolicy{Name: restartPolicyName}
	}
}

func handleContainerErr(err error, m string, args ...interface{}) {
	if err != nil {
		log.Errorf(err, m, args...)
		Stop()
		panic(err)
	}
}
