// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package utils

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"k8s-connectors/connectors/yos/pkg/config"
)

type AwsSdkProvider = func(ctx context.Context, key string, secret string) (*s3.S3, error)

func NewDefaultProvider() AwsSdkProvider {
	return func(ctx context.Context, key string, secret string) (*s3.S3, error) {
		ses, err := session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(key, secret, ""),
			Endpoint:    aws.String(config.Endpoint),
			EndpointResolver: endpoints.ResolverFunc(
				func(service, region string, opts ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
					return endpoints.ResolvedEndpoint{URL: config.Endpoint}, nil
				}),
			Region:           aws.String(config.AwsRegion),
			S3ForcePathStyle: aws.Bool(true),
		})

		if err != nil {
			return nil, fmt.Errorf("unable to get yos sdk: %v", err)
		}

		return s3.New(ses), nil
	}
}
