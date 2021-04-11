package main_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/sethvargo/go-retry"
	"github.com/tclemos/goit"
	"github.com/tclemos/goit/dockerfile"
)

const (
	host = "0.0.0.0"
	port = "8080"
)

func TestMain(m *testing.M) {

	ctx := context.Background()

	pb := map[docker.Port][]docker.PortBinding{
		docker.Port(fmt.Sprintf("%s/tcp", port)): {{HostIP: host, HostPort: port}},
	}

	// Prepare container
	c := dockerfile.NewContainer(dockerfile.Params{
		// specify here the container name
		ContainerName: "myapp",

		// specify here the dockerfile path
		DockerFilePath: "./Dockerfile",

		// use this to specify the arguments the dockerfile requires to work properly
		BuildArgs: []docker.BuildArg{{
			Name: "HOST", Value: host,
		}},

		// use this to set the environment variables for your container
		Env: map[string]string{
			"MYAPP_PORT": port,
		},

		// define the port bindings to open external ports to the host
		PortBindings: pb,

		// use the AfterStart function to make sure your container is ready for test,
		// for example, make a request to a known URL of your service to make sure
		// it's up and running with a retry logic, if it fails, return
		// an error to stop de test pipeline
		AfterStart: func(c context.Context, r *dockertest.Resource, m *map[string]interface{}) error {

			b, _ := retry.NewFibonacci(500 * time.Millisecond)
			b = retry.WithMaxRetries(10, b)
			b = retry.WithCappedDuration(20*time.Second, b)

			err := retry.Do(c, b, func(ctx context.Context) error {
				addr := fmt.Sprintf("http://localhost:%s/ping", port)
				res, err := http.DefaultClient.Get(addr)
				if err != nil || res.StatusCode != http.StatusOK {
					fmt.Println("waiting on application to initialize...")
					return retry.RetryableError(err)
				}
				return nil
			})
			if err != nil {
				return err
			}

			return nil
		},
	})

	// Start container
	goit.Start(ctx, c)

	// Run tests
	code := m.Run()

	// Stop containers
	goit.Stop()

	// finalize test execution
	os.Exit(code)
}

func TestFoo(t *testing.T) {

	addr := fmt.Sprintf("http://localhost:%s/foo", port)
	res, err := http.DefaultClient.Get(addr)
	if err != nil || res.StatusCode != http.StatusOK {
		t.Error("Failed to get foo API")
	}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		t.Error("Failed to read foo response body")
	}

	bodyText := string(bodyBytes)
	if bodyText != "bar" {
		t.Errorf("Invalid response body for foo API, expected: bar, found: %s", bodyText)
	}
}
