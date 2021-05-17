package kafka_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/tclemos/goit"
	"github.com/tclemos/goit/kafka"
)

const (
	Region    = "eu-central-1"
	Port      = 4566
	QueueName = "example_queue"
)

var c *kafka.Container

func TestMain(m *testing.M) {

	ctx := context.Background()

	// Prepare container
	c = kafka.NewContainer(kafka.Params{
		Topics: []string{
			"TopicOne",
			"TopicTwo",
		},
	})

	// Start container
	opt := goit.DefaultOptions()
	opt.AutoRemoveContainers = true
	goit.StartWithOptions(ctx, opt, c)

	// Run tests
	code := m.Run()

	// Stop containers
	goit.Stop()

	// finalize test execution
	os.Exit(code)
}

func TestKafka(t *testing.T) {

	type thing struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}

	sample := thing{
		Id:   1,
		Name: "something",
	}

	b, err := json.Marshal(sample)
	if err != nil {
		t.Errorf("Unable to marshal thing: %v", err)
		return
	}

	c.Producer.Produce("TopicOne", b)

	message, err := c.Consumers["TopicOne"].Consume()
	if err != nil {
		t.Errorf("Failed to receive messages: %v", err)
		return
	}

	var th thing
	if err := json.Unmarshal(message, &th); err != nil {
		t.Errorf("Failed to unmarshal message body: %v", err)
		return
	}

	if th.Id != 1 {
		t.Errorf("Invalid Id, expected 1, found: %d", th.Id)
		return
	}

	if th.Name != "something" {
		t.Errorf("Invalid name, expected something, found: %s", th.Name)
		return
	}
}
