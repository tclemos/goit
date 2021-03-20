package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/tclemos/goit/log"
)

var (
	visibilityTimeout = int64(60)
)

type SqsService struct {
	svc *sqs.SQS
}

func newSqsService(s *session.Session) *SqsService {
	svc := sqs.New(s)

	return &SqsService{
		svc: svc,
	}
}

func (s SqsService) Send(qn string, m *sqs.Message) error {
	urlOutput, err := s.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &qn,
	})
	if err != nil {
		log.Error(err, "failed to get the queueUrl for queue: %s", qn)
		return nil
	}

	_, err = s.svc.SendMessage(&sqs.SendMessageInput{
		MessageAttributes: m.MessageAttributes,
		MessageBody:       m.Body,
		QueueUrl:          urlOutput.QueueUrl,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s SqsService) Receive(qn string) ([]*sqs.Message, error) {
	urlOutput, err := s.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &qn,
	})
	if err != nil {
		log.Error(err, "failed to get the queueUrl for queue: %s", qn)
		return nil, err
	}

	msgResult, err := s.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            urlOutput.QueueUrl,
		MaxNumberOfMessages: aws.Int64(1),
		VisibilityTimeout:   &visibilityTimeout,
	})
	if err != nil {
		return nil, err
	}

	return msgResult.Messages, nil
}
