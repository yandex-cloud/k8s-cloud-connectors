// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gin-gonic/gin"

	"example/pkg/awscompatibility"
	"example/pkg/util"
)

func main() {
	key, secret := util.GetEnvOrDie("AWS_ACCESS_KEY_ID"), util.GetEnvOrDie("AWS_SECRET_ACCESS_KEY")

	ymq, err := awscompatibility.NewSQSClient(context.TODO(), credentials.NewStaticCredentials(key, secret, ""))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("ymq client with %s/%s built\n", key, secret)

	ymqURL := util.StrPtr(util.GetEnvOrDie("YMQ_URL"))
	log.Printf("ymq url is %s\n", *ymqURL)

	r := gin.Default()
	r.POST("/report", func(c *gin.Context) {
		if _, err := ymq.SendMessage(&sqs.SendMessageInput{
			MessageBody: util.StrPtr(c.Query("message")),
			QueueUrl:    ymqURL,
		}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
