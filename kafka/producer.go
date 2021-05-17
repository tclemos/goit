package kafka

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/tclemos/goit/log"
)

type Producer struct {
	p *kafka.Producer
}

func newProducer(cm *kafka.ConfigMap) (*Producer, error) {
	p, err := kafka.NewProducer(cm)

	if err != nil {
		return nil, err
	}

	return &Producer{
		p: p,
	}, err
}

func (p *Producer) Produce(topic string, message []byte) error {

	c := make(chan kafka.Event, 1000000)

	err := p.p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
	}, c)

	if err != nil {
		log.Errorf(err, "Failed to produce. topic: %s, message: %s", topic, message)
		return err
	}

	r := <-c

	if r != nil {
		return fmt.Errorf(r.String())
	}

	return nil
}
