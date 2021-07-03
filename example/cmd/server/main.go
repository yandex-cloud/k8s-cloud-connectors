// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gin-gonic/gin"
)

func getEnvOrDie(key string) string {
	res, ok := os.LookupEnv(key)
	if !ok {
		log.Fatal("unable to get environmental variable " + key)
	}
	return res
}

const (
	SqsEndpoint = "message-queue.api.cloud.yandex.net"
	SqsRegion   = "ru-central1"
)

func newSQSClient(_ context.Context, cred *credentials.Credentials) (*sqs.SQS, error) {
	ses, err := session.NewSession(
		&aws.Config{
			Credentials: cred,
			Endpoint:    aws.String(SqsEndpoint),
			EndpointResolver: endpoints.ResolverFunc(
				func(service, region string, opts ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
					return endpoints.ResolvedEndpoint{URL: SqsEndpoint}, nil
				},
			),
			Region:           aws.String(SqsRegion),
			S3ForcePathStyle: aws.Bool(true),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("unable to get sqs sdk: %w", err)
	}

	return sqs.New(ses), nil
}

func strPtr(str string) *string {
	return &str
}

func main() {
	key, secret := getEnvOrDie("AWS_ACCESS_KEY_ID"), getEnvOrDie("AWS_SECRET_ACCESS_KEY")

	ymq, err := newSQSClient(context.TODO(), credentials.NewStaticCredentials(key, secret, ""))
	if err != nil {
		log.Fatal(err)
	}

	ymqURL := strPtr(getEnvOrDie("YMQ_URL"))

	r := gin.Default()
	r.POST("/report", func(c *gin.Context) {
		if _, err := ymq.SendMessage(&sqs.SendMessageInput{
			MessageBody: strPtr(c.Query("image")),
			QueueUrl:    ymqURL,
		}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "malformed or empty string",
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
