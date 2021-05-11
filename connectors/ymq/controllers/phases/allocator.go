// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	connectorsv1 "k8s-connectors/connectors/ymq/api/v1"
	"k8s-connectors/connectors/ymq/controllers/adapter"
	ymqutils "k8s-connectors/connectors/ymq/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceAllocator struct {
	Client *client.Client
	Sdk    adapter.YandexMessageQueueAdapter
}

func (r *ResourceAllocator) IsUpdated(ctx context.Context, resource *connectorsv1.YandexMessageQueue) (bool, error) {
	key, secret, err := ymqutils.KeyAndSecretFromStaticAccessKey(ctx, resource, *r.Client)
	if err != nil {
		return false, err
	}
	lst, err := r.Sdk.List(ctx, key, secret)
	if err != nil {
		return false, err
	}
	for _, queue := range lst {
		if *queue == resource.Status.QueueURL {
			return true, nil
		}
	}
	return false, nil
}

func (r *ResourceAllocator) Update(ctx context.Context, log logr.Logger, resource *connectorsv1.YandexMessageQueue) error {
	key, secret, err := ymqutils.KeyAndSecretFromStaticAccessKey(ctx, resource, *r.Client)
	if err != nil {
		return err
	}
	attributes := make(map[string]*string)
	res, err := r.Sdk.Create(ctx, key, secret, attributes, resource.Spec.Name)
	if err != nil {
		return err
	}

	resource.Status.QueueURL = res
	if err := (*r.Client).Status().Update(ctx, resource); err != nil {
		return fmt.Errorf("error while creating resource: %v", err)
	}

	log.Info("resource successfully allocated")
	return nil
}

func (r *ResourceAllocator) Cleanup(ctx context.Context, log logr.Logger, resource *connectorsv1.YandexMessageQueue) error {
	key, secret, err := ymqutils.KeyAndSecretFromStaticAccessKey(ctx, resource, *r.Client)
	if err != nil {
		return err
	}

	err = r.Sdk.Delete(ctx, key, secret, resource.Status.QueueURL)
	if err != nil {
		return err
	}

	log.Info("resource successfully deleted")
	return nil
}
