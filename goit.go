package goit

import (
	"context"

	"github.com/ory/dockertest/v3"
	"github.com/tclemos/goit/log"
)

var (
	pool      *dockertest.Pool
	resources []*dockertest.Resource
)

// Start the integration test environment
func Start(containers ...Container) {
	log.Log("initializing containers")
	resources = []*dockertest.Resource{}
	Ctx = context.WithValue(context.Background(), valuesKey, map[string]interface{}{})

	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Errorf(err, "failed to create docker pool")
		panic(err)
	}

	for _, c := range containers {
		o, err := c.Options()
		log.Log("loading container with options: %v", o)
		handleContainerErr(err, "can't load container")

		r, err := startContainer(Ctx, pool, o)
		handleContainerErr(err, "can't start container")

		log.Log("executing AfterStart")
		err = c.AfterStart(Ctx, r)
		handleContainerErr(err, "failed to execute AfterStarted")

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

func handleContainerErr(err error, m string, args ...interface{}) {
	if err != nil {
		log.Errorf(err, m, args)
		Stop()
		panic(err)
	}
}
