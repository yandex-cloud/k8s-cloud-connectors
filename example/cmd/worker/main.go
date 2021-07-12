// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"

	"example/pkg/awscompatibility"
)

func getEnvOrDie(key string) string {
	res, ok := os.LookupEnv(key)
	if !ok {
		log.Fatal("unable to get environmental variable \"" + key + "\"")
	}
	return res
}

func processMessage(client *s3.S3, s3URL, msg string) error {
	var unencoded map[string]string
	if err := json.Unmarshal([]byte(msg), &unencoded); err != nil {
		return err
	}

	filename := unencoded["filename"]

	_, err := client.PutObject(
		&s3.PutObjectInput{
			Body:   bytes.NewReader([]byte("time: " + time.Now().String() + "\nmessage: " + unencoded["content"])),
			Bucket: &s3URL,
			Key:    &filename,
		},
	)
	return err
}

// main function of Worker infinitely polls YMQ specified by environmental variables and processes
// queries placed there
func main() {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	key, secret := getEnvOrDie("AWS_ACCESS_KEY_ID"), getEnvOrDie("AWS_SECRET_ACCESS_KEY")

	ymq, err := awscompatibility.NewSQSClient(ctx, credentials.NewStaticCredentials(key, secret, ""))
	if err != nil {
		log.Fatal(err)
	}

	ymqURL := getEnvOrDie("YMQ_URL")

	s3Client, err := awscompatibility.NewS3Client(ctx, credentials.NewStaticCredentials(key, secret, ""))
	if err != nil {
		log.Fatal(err)
	}

	s3Name := getEnvOrDie("S3_URL")

	for {
		numberOfMessages := int64(10)
		messages, err := ymq.ReceiveMessage(
			&sqs.ReceiveMessageInput{
				AttributeNames: []*string{
					aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
				},
				MessageAttributeNames: []*string{
					aws.String(sqs.QueueAttributeNameAll),
				},
				MaxNumberOfMessages: &numberOfMessages,
				QueueUrl:            &ymqURL,
			},
		)
		if err != nil {
			log.Print(err)
			continue
		}

		for _, msg := range messages.Messages {
			if err := processMessage(s3Client, s3Name, *msg.Body); err != nil {
				log.Print(err)
				continue
			}
			_, err = ymq.DeleteMessage(
				&sqs.DeleteMessageInput{
					QueueUrl:      &ymqURL,
					ReceiptHandle: msg.ReceiptHandle,
				},
			)
			if err != nil {
				log.Print(err)
				continue
			}
		}
	}
}
