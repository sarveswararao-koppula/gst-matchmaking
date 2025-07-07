package queue

import (
	"errors"
	"mm/properties"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var qURL string = properties.Prop.AWS_SQS_URL
var qCredPath string = properties.Prop.AWS_SQS_CRED_PATH

func getSess() *session.Session {

	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewSharedCredentials(qCredPath, "default"),
	})

	return sess
}

var sess *session.Session = getSess()
var svc *sqs.SQS = sqs.New(sess)

//Send ...
func Send(enqData map[string]string) (string, error) {

	if enqData["publisher"] == "" {
		return "", errors.New("publisher missing")
	}

	if enqData["jsonDataStr"] == "" {
		return "", errors.New("jsonDataStr missing")
	}

	result, err := svc.SendMessage(&sqs.SendMessageInput{
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"publisher": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(enqData["publisher"]),
			},
			"jsonDataStr": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(enqData["jsonDataStr"]),
			},
		},
		MessageBody:            aws.String(enqData["msgBody"]),
		// MessageDeduplicationId: aws.String(enqData["msgDuplicationID"]),
		// MessageGroupId:         aws.String(enqData["msgGroupID"]),
		QueueUrl:               &qURL,
	})

	if err != nil {
		return "", err
	}

	return *result.MessageId, nil
}

//Receive ...
func Receive() (map[string]string, *string, string, error) {

	deqData := make(map[string]string)

	result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            &qURL,
		MaxNumberOfMessages: aws.Int64(1),
		VisibilityTimeout:   aws.Int64(60),
		WaitTimeSeconds:     aws.Int64(0),
	})

	if err != nil {
		return deqData, nil, "", err
	}

	if len(result.Messages) == 0 {
		return deqData, nil, "", errors.New("queue empty")
	}

	rHandle := result.Messages[0].ReceiptHandle
	msgID := *result.Messages[0].MessageId

	for k, v := range result.Messages[0].MessageAttributes {
		deqData[k] = *v.StringValue
	}

	return deqData, rHandle, msgID, nil
}

//Delete ...
func Delete(messageHandle *string) error {

	if *messageHandle == "" {
		return errors.New("ReceiptHandle Empty")
	}

	_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &qURL,
		ReceiptHandle: messageHandle,
	})

	return err
}
