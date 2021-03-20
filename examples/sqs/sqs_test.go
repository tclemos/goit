package sqs_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/tclemos/goit"
	"github.com/tclemos/goit/aws"
)

const (
	Region    = "eu-central-1"
	Port      = 4566
	QueueName = "example_queue"
)

var sqsService *aws.SqsService

func TestMain(m *testing.M) {

	ctx := context.Background()

	// Prepare container
	c := aws.NewContainer(aws.Params{
		Region: Region,
		Port:   Port,
		SqsQueues: []aws.SqsQueue{
			{Name: QueueName},
		},
	})

	// Start container
	goit.Start(ctx, c)
	sqsService = c.SqsService

	// Run tests
	code := m.Run()

	// Stop containers
	goit.Stop()

	// finalize test execution
	os.Exit(code)
}

func TestSqs(t *testing.T) {

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

	body := string(b)

	sqsService.Send(QueueName, &sqs.Message{
		Body: &body,
	})

	messages, err := sqsService.Receive(QueueName)
	if err != nil {
		t.Errorf("Failed to receive messages: %v", err)
		return
	}

	count := len(messages)
	if count != 1 {
		t.Errorf("Invalid count, expected 1, found: %d", count)
		return
	}

	message := messages[0]
	var th thing
	if err := json.Unmarshal([]byte(*message.Body), &th); err != nil {
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
