// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phase

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s-connectors/pkg/configmap"
)

func ProvideConfigmap(
	ctx context.Context,
	cl client.Client,
	log logr.Logger,
	objectName, kindName, namespace string,
	contents map[string]string,
) error {
	log.V(1).Info("started")

	exists, err := configmap.Exists(ctx, cl, objectName, namespace, kindName)
	if err != nil {
		return fmt.Errorf("unable to check configmap existence: %v", err)
	}
	if exists {
		return nil
	}

	if err := configmap.Put(ctx, cl, objectName, namespace, kindName, contents); err != nil {
		return err
	}

	log.Info("successful")
	return nil
}

func RemoveConfigmap(
	ctx context.Context,
	cl client.Client,
	log logr.Logger,
	objectName, kindName, namespace string,
) error {
	log.V(1).Info("started")

	if err := configmap.Remove(ctx, cl, objectName, namespace, kindName); err != nil {
		return fmt.Errorf("unable to remove configmap: %v", err)
	}

	log.Info("successful")
	return nil
}
