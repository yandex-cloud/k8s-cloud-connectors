// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package main

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"

	"example/pkg/awscompatibility"
	"example/pkg/util"
)

func processMessage(client *s3.S3, s3URL, msg *string) error {
	_, err := client.PutObject(&s3.PutObjectInput{
		Body:                      bytes.NewReader([]byte("message: " + *msg + "time: " + time.Now().String())),
		Bucket:                    s3URL,
		Key:                       util.StrPtr("msg.txt"),
	})
	return err
}

func main() {
	key, secret := util.GetEnvOrDie("AWS_ACCESS_KEY_ID"), util.GetEnvOrDie("AWS_SECRET_ACCESS_KEY")

	ymq, err := awscompatibility.NewSQSClient(context.TODO(), credentials.NewStaticCredentials(key, secret, ""))
	if err != nil {
		log.Fatal(err)
	}

	ymqURL := util.StrPtr(util.GetEnvOrDie("YMQ_URL"))

	s3Client, err := awscompatibility.NewS3Client(context.TODO(), credentials.NewStaticCredentials(key, secret, ""))
	if err != nil {
		log.Fatal(err)
	}

	s3Name := util.StrPtr(util.GetEnvOrDie("S3_URL"))

	for {
		messages, err := ymq.ReceiveMessage(&sqs.ReceiveMessageInput{
			AttributeNames: []*string{
				aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
			},
			MessageAttributeNames: []*string{
				aws.String(sqs.QueueAttributeNameAll),
			},
			MaxNumberOfMessages: util.Int64Ptr(10),
			QueueUrl:            ymqURL,
		})
		if err != nil {
			log.Print(err)
			continue
		}

		for _, msg := range messages.Messages {
			if err := processMessage(s3Client, s3Name, msg.Body); err != nil {
				log.Print(err)
				continue
			}
			_, err = ymq.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      ymqURL,
				ReceiptHandle: msg.ReceiptHandle,
			})
			if err != nil {
				log.Print(err)
				continue
			}
		}
	}
}
