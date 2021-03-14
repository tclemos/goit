package localstack

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/tclemos/goit/log"
)

var (
	visibilityTimeout = int64(60)
)

type SqsReceiver struct {
	sqs       *sqs.SQS
	queueName string
	queueURL  string
	session   *session.Session
}

func NewSqsReceiver(qn string, s *session.Session) *SqsReceiver {
	svc := sqs.New(s)
	urlOutput, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &qn,
	})

	if err != nil {
		log.Error(err, "failed to get the queueUrl for queue: %s", qn)
		panic(err)
	}

	return &SqsReceiver{
		sqs:       svc,
		queueName: qn,
		queueURL:  *urlOutput.QueueUrl,
		session:   s,
	}
}

func (r SqsReceiver) Receive() ([]*sqs.Message, error) {
	msgResult, err := r.sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            &r.queueURL,
		MaxNumberOfMessages: aws.Int64(1),
		VisibilityTimeout:   &visibilityTimeout,
	})
	if err != nil {
		return nil, err
	}

	return msgResult.Messages, nil
}
