package postgres

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/sethvargo/go-retry"
	"github.com/tclemos/goit"
	"github.com/tclemos/goit/log"

	// packages required to execute migrations
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	port = 5432
)

// Params needed to start a postgres container
type Params struct {
	goit.ContainerParams
	Port     int
	User     string
	Password string
	Database string
}

// Container metadata to load a container for postgres database
type Container struct {
	params Params
	url    url.URL
}

// NewContainer creates a new instance of Container
func NewContainer(p Params) *Container {
	return &Container{
		params: p,
	}
}

// Options to start a postgres container accordingly to the params
func (c *Container) Options() (*dockertest.RunOptions, error) {
	p := c.params.Port
	if p == 0 {
		p = port
	}

	strPort := strconv.Itoa(p)
	pb := map[docker.Port][]docker.PortBinding{}
	pb[docker.Port(fmt.Sprintf("%d/tcp", port))] = []docker.PortBinding{{
		HostIP:   "0.0.0.0",
		HostPort: strPort,
	}}

	repo, tag := c.params.GetRepoTag("postgres", "latest")
	env := c.params.MergeEnv([]string{
		"POSTGRES_DB=" + c.params.Database,
		"POSTGRES_USER=" + c.params.User,
		"POSTGRES_PASSWORD=" + c.params.Password,
		"POSTGRES_HOST_AUTH_METHOD=trust",
	})

	return &dockertest.RunOptions{
		Repository:   repo,
		Tag:          tag,
		Env:          env,
		PortBindings: pb,
	}, nil
}

// AfterStart will check the connection and execute migrations
func (c *Container) AfterStart(ctx context.Context, r *dockertest.Resource) error {
	// db url
	c.url = c.createDBURL(r)

	// check db connection
	err := c.checkDb(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *Container) Url() url.URL {
	return c.url
}

func (c *Container) createDBURL(r *dockertest.Resource) url.URL {
	// find db host
	id := fmt.Sprintf("%d/tcp", port)
	h := r.GetBoundIP(id)
	p := r.GetPort(id)
	host := net.JoinHostPort(h, p)

	// Build the connection URL.
	dbURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.params.User, c.params.Password),
		Host:   host,
		Path:   c.params.Database,
	}
	q := dbURL.Query()
	q.Add("sslmode", "disable")
	dbURL.RawQuery = q.Encode()
	return dbURL
}

func (c *Container) checkDb(ctx context.Context) error {
	log.Logf("checking postgres connection at %s", c.url.String())
	// prepare a connection verification interval. Use a Fibonacci backoff
	// instead of exponential so wait times scale appropriately.
	b, err := retry.NewFibonacci(500 * time.Millisecond)
	if err != nil {
		log.Error(err, "failed to configure retries to check db connection")
		return err
	}

	b = retry.WithMaxRetries(10, b)
	b = retry.WithCappedDuration(10*time.Second, b)

	// Establish a connection to the database.
	err = retry.Do(ctx, b, func(ctx context.Context) error {
		_, err := pgxpool.Connect(ctx, c.url.String())
		if err != nil {
			log.Log("waiting on postgres server to be available")
			return retry.RetryableError(err)
		}
		return nil
	})
	if err != nil {
		log.Error(err, "failed to start postgres")
		return err
	}

	log.Logf("postgres available at: %s", c.url.String())
	return nil
}
