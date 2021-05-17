package kafka

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/tclemos/goit"
	"github.com/tclemos/goit/log"
)

const (
	brokerPort = 9092
	clientPort = 9093
)

// Params needed to start a kafka container
type Params struct {
	goit.ContainerParams
	BrokerPort int
	ClientId   string
	ClientPort int
	Topics     []string
}

// Container metadata to load a container for kafka
type Container struct {
	params    Params
	Producer  *Producer
	Consumers map[string]*Consumer
}

// NewContainer creates a new instance of Container
func NewContainer(p Params) *Container {
	return &Container{
		params: p,
	}
}

// Options to start a kafka container accordingly to the params
func (c *Container) Options() (*dockertest.RunOptions, error) {

	strPort := strconv.Itoa(c.getBrokerPort())
	pb := map[docker.Port][]docker.PortBinding{}
	pb[docker.Port(fmt.Sprintf("%d/tcp", brokerPort))] = []docker.PortBinding{{
		HostIP:   "0.0.0.0",
		HostPort: strPort,
	}}

	strPort = strconv.Itoa(c.getClientPort())
	pb[docker.Port(fmt.Sprintf("%d/tcp", clientPort))] = []docker.PortBinding{{
		HostIP:   "0.0.0.0",
		HostPort: strPort,
	}}

	repo, tag := c.params.GetRepoTag("confluentinc/cp-kafka", "5.3.0")
	env := c.params.MergeEnv([]string{
		"KAFKA_BROKER_ID=1",
		fmt.Sprintf("KAFKA_LISTENERS=\"PLAINTEXT://0.0.0.0:%d,BROKER://0.0.0.0:%d", c.getClientPort(), c.getBrokerPort()),
		"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=\"BROKER:PLAINTEXT,PLAINTEXT:PLAINTEXT\"",
		"KAFKA_INTER_BROKER_LISTENER_NAME=BROKER",
		"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1",
	})

	return &dockertest.RunOptions{
		Repository:   repo,
		Tag:          tag,
		Env:          env,
		PortBindings: pb,
	}, nil
}

// AfterStart
func (c *Container) AfterStart(ctx context.Context, r *dockertest.Resource) error {

	id := fmt.Sprintf("%d/tcp", clientPort)
	h := r.GetBoundIP(id)
	p := r.GetPort(id)
	host := net.JoinHostPort(h, p)

	clientId := uuid.New().String()
	if len(c.params.ClientId) > 0 {
		clientId = c.params.ClientId
	}

	cm := &kafka.ConfigMap{
		"bootstrap.servers":     host,
		"broker.address.family": "v4",
		"client.id":             clientId,
	}

	ac, err := kafka.NewAdminClient(cm)
	if err != nil {
		log.Error(err, "failed to configure retries to check db connection")
		return err
	}
	defer ac.Close()

	var ts []kafka.TopicSpecification

	for _, topic := range c.params.Topics {
		ts = append(ts, kafka.TopicSpecification{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})

		ccm := *cm
		ccm.Set(fmt.Sprintf("group.id=%s", topic))

		consumer, err := newConsumer(&ccm)
		if err != nil {
			log.Errorf(err, "failed to create a consumer for topic %s", topic)
			return err
		}

		c.Consumers[topic] = consumer
	}

	_, err = ac.CreateTopics(
		ctx,
		ts,
		nil)

	if err != nil {
		log.Error(err, "failed to create kafka topics")
		return err
	}

	c.Producer, err = newProducer(cm)
	if err != nil {
		log.Error(err, "failed to create kafka producer")
		return err
	}

	return nil
}

func (c *Container) getBrokerPort() int {
	bp := c.params.BrokerPort
	if bp == 0 {
		bp = brokerPort
	}

	return bp
}

func (c *Container) getClientPort() int {
	cp := c.params.ClientPort
	if cp == 0 {
		cp = clientPort
	}
	return cp
}
