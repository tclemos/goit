package aws

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
	"github.com/tclemos/go-dockertest-example/e2e"
	"github.com/tclemos/goit/log"
)

const (
	edgePort = 4566
)

type SqsQueue struct {
	Name string
}

// Params needed to start a aws container
type Params struct {
	Region    string
	Port      int
	SqsQueues []SqsQueue
}

// Container metadata to load a container for aws environment
type Container struct {
	params Params
}

// NewContainer creates a new instance of Container
func NewContainer(p Params) *Container {
	return &Container{
		params: p,
	}
}

// Options to start a localstack container accordingly to the params
func (c *Container) Options() (*dockertest.RunOptions, error) {
	pb := map[docker.Port][]docker.PortBinding{}
	pb[docker.Port(fmt.Sprintf("%d/tcp", edgePort))] = []docker.PortBinding{{HostIP: "0.0.0.0", HostPort: strconv.Itoa(edgePort)}}

	return &dockertest.RunOptions{
		Repository: "localstack/localstack",
		Tag:        "latest",
		Env: []string{
			"SERVICES=sqs",
			"DATA_DIR=/tmp/localstack/data",
		},
		PortBindings: pb,
	}, nil
}

// AfterStart will wait until the container is ready to be consumed
func (c *Container) AfterStart(ctx context.Context, r *dockertest.Resource) error {

	// sets the endpoint to aws config
	awsconfig := e2e.GetValue("awsconfig")
	if awsconfig == nil {
		awsconfig = aws.NewConfig()
	}

	awsconfig.(*aws.Config).
		WithEndpoint(fmt.Sprintf("http://localhost:%d", edgePort)).
		WithCredentialsChainVerboseErrors(true).
		WithHTTPClient(&http.Client{Timeout: 10 * time.Second}).
		WithMaxRetries(2).
		WithCredentials(credentials.NewStaticCredentials("foo", "bar", "")).
		WithRegion(c.params.Region).
		WithDisableSSL(true)

	e2e.AddValue("awsconfig", awsconfig)

	s, err := session.NewSession(awsconfig.(*aws.Config))
	if err != nil {
		log.Errorf(err, "failed to create aws session")
		return err
	}
	svc := sqs.New(s)

	// await initialization
	c.awaitInitialization(ctx, svc)

	// create sqs queues
	for _, q := range c.params.SqsQueues {
		_, err := svc.CreateQueue(&sqs.CreateQueueInput{
			QueueName: aws.String(q.Name),
		})
		if err != nil {
			log.Errorf(err, "failed to create queue: %s", q.Name)
			return err
		}
	}

	return nil
}

func (c *Container) awaitInitialization(ctx context.Context, svc *sqs.SQS) error {
	// prepare a connection verification interval. Use a Fibonacci backoff
	// instead of exponential so wait times scale appropriately.
	b, err := retry.NewFibonacci(500 * time.Millisecond)
	if err != nil {
		err = errors.Wrap(err, "failed to configure retries to wait initialization")
		return err
	}

	b = retry.WithMaxRetries(10, b)
	b = retry.WithCappedDuration(100*time.Second, b)

	// Tries to create a queue to make sure the localstack is up.
	var createQueue *sqs.CreateQueueOutput
	err = retry.Do(ctx, b, func(ctx context.Context) error {
		createQueue, err = svc.CreateQueue(&sqs.CreateQueueInput{
			QueueName: aws.String("test-Resource"),
		})
		if err != nil {
			log.Log("waiting aws server to initialize...")
			return retry.RetryableError(err)
		}
		log.Log("aws server initialized")
		return nil
	})
	if err != nil {
		log.Error(err, "failed to initialize aws server")
		return err
	}

	if _, err := svc.DeleteQueue(&sqs.DeleteQueueInput{
		QueueUrl: createQueue.QueueUrl,
	}); err != nil {
		return err
	}

	return nil
}

func CreateConfig(port int, region string) *aws.Config {
	return aws.NewConfig().
		WithEndpoint(fmt.Sprintf("http://localhost:%d", port)).
		WithCredentialsChainVerboseErrors(true).
		WithHTTPClient(&http.Client{Timeout: 10 * time.Second}).
		WithMaxRetries(2).
		WithCredentials(credentials.NewStaticCredentials("foo", "bar", "")).
		WithRegion(region).
		WithDisableSSL(true)
}
