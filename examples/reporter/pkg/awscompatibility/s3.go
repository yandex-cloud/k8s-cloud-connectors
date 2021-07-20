// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package awscompatibility

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func NewS3Client(_ context.Context, cred *credentials.Credentials) (*s3.S3, error) {
	ses, err := session.NewSession(
		&aws.Config{
			Credentials: cred,
			Endpoint:    aws.String(S3Endpoint),
			EndpointResolver: endpoints.ResolverFunc(
				func(service, region string, opts ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
					return endpoints.ResolvedEndpoint{URL: S3Endpoint}, nil
				},
			),
			Region:           aws.String(S3Region),
			S3ForcePathStyle: aws.Bool(true),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("unable to get yos sdk: %w", err)
	}

	return s3.New(ses), nil
}
