// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package awsutils

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sakey "k8s-connectors/connector/sakey/api/v1"
)

func CredentialsFromStaticAccessKey(
	ctx context.Context, namespace, sakeyName string, cl client.Client,
) (*credentials.Credentials, error) {
	var key sakey.StaticAccessKey
	if err := cl.Get(
		ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      sakeyName,
		}, &key,
	); err != nil {
		return nil, fmt.Errorf("unable to retrieve corresponding SAKey: %w", err)
	}

	var secret v1.Secret
	if err := cl.Get(
		ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      key.Status.SecretName,
		}, &secret,
	); err != nil {
		return nil, fmt.Errorf("unable to retrieve corresponding secret: %w", err)
	}

	return credentials.NewStaticCredentials(string(secret.Data["key"]), string(secret.Data["secret"]), ""), nil
}
