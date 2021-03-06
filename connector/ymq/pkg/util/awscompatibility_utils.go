// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package util

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/yandex-cloud/k8s-cloud-connectors/connector/ymq/pkg/config"
)

func NewSQSClient(_ context.Context, cred *credentials.Credentials) (*sqs.SQS, error) {
	ses, err := session.NewSession(
		&aws.Config{
			Credentials: cred,
			Endpoint:    aws.String(config.Endpoint),
			EndpointResolver: endpoints.ResolverFunc(
				func(service, region string, opts ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
					return endpoints.ResolvedEndpoint{URL: config.Endpoint}, nil
				},
			),
			Region:           aws.String(config.AWSRegion),
			S3ForcePathStyle: aws.Bool(true),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("unable to get %s sdk: %w", config.ShortName, err)
	}

	return sqs.New(ses), nil
}
