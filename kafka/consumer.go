package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/tclemos/goit/log"
)

type Consumer struct {
	c *kafka.Consumer
}

func newConsumer(cm *kafka.ConfigMap) (*Consumer, error) {
	c, err := kafka.NewConsumer(cm)

	if err != nil {
		return nil, err
	}

	return &Consumer{
		c: c,
	}, err
}

func (c *Consumer) Consume() ([]byte, error) {
	m, err := c.c.ReadMessage(-1)
	if err != nil {
		log.Error(err, "error consuming message.")
		return []byte{}, err
	}

	return m.Value, nil
}
