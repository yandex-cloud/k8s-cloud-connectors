// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package util

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
	"k8s-connectors/connectors/sakey/controllers/adapter"
	sakeyconfig "k8s-connectors/connectors/sakey/pkg/config"
	"k8s-connectors/pkg/errors"
)

func GetStaticAccessKey(ctx context.Context, keyID string, saID string, clusterName string, name string, adapter adapter.StaticAccessKeyAdapter) (*awscompatibility.AccessKey, error) {
	if keyID != "" {
		res, err := adapter.Read(ctx, keyID)
		if err != nil {
			// If resource was not found then it does not exist,
			// but this error is not fatal, just a mismatch between
			// out status and real world state.
			if errors.CheckRPCErrorNotFound(err) {
				return nil, nil
			}
			// Otherwise, it is fatal
			return nil, fmt.Errorf("cannot get resource from cloud: %v", err)
		}

		// Everything is fine, we have found it
		return res, nil
	}

	// We may have not yet written this key into status,
	// But we can list objects and match by description
	// TODO (covariance) pagination
	lst, err := adapter.List(ctx, saID)
	if err != nil {
		return nil, fmt.Errorf("cannot list resources in cloud: %v", err)
	}

	for _, res := range lst {
		if res.Description == sakeyconfig.GetStaticAccessKeyDescription(clusterName, name) {
			// By description match we deduce that its our key
			return res, nil
		}
	}

	return nil, nil
}
