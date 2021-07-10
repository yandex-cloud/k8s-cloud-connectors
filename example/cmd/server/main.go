// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws/credentials"
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

func composeMessage(filename, contents string) *string {
	res, _ := json.Marshal(map[string]string{
		"filename" : filename,
		"content" : contents,
	})

	strRes := string(res)

	return &strRes
}

func main() {
	key, secret := getEnvOrDie("AWS_ACCESS_KEY_ID"), getEnvOrDie("AWS_SECRET_ACCESS_KEY")

	ymq, err := awscompatibility.NewSQSClient(context.TODO(), credentials.NewStaticCredentials(key, secret, ""))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("ymq client with %s/%s built\n", key, secret)

	ymqURL := getEnvOrDie("YMQ_URL")
	log.Printf("ymq url is %s\n", ymqURL)

	http.HandleFunc(
		"/report", func(w http.ResponseWriter, req *http.Request) {
			resp := map[string]string{
				"message": "ok",
			}

			defer func() { _ = req.Body.Close() }()

			defer func() {
				bytes, _ := json.Marshal(resp)
				_, _ = w.Write(bytes)
			}()

			filename, ok := req.URL.Query()["filename"]
			if !ok || len(filename) == 0 {
				resp["message"] = "no filename provided in query"
				return
			}

			msg, err := ioutil.ReadAll(req.Body)
			if err != nil {
				resp["message"] = err.Error()
				return
			}

			if _, err := ymq.SendMessage(
				&sqs.SendMessageInput{
					MessageBody: composeMessage(filename[0], string(msg)),
					QueueUrl:    &ymqURL,
				},
			); err != nil {
				resp["message"] = err.Error()
				return
			}
		},
	)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
