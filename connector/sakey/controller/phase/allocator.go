// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phase

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "k8s-connectors/connector/sakey/api/v1"
	"k8s-connectors/connector/sakey/controller/adapter"
	sakeyconfig "k8s-connectors/connector/sakey/pkg/config"
	sakeyutils "k8s-connectors/connector/sakey/pkg/util"
	"k8s-connectors/pkg/secret"
)

type Allocator struct {
	Sdk       adapter.StaticAccessKeyAdapter
	Client    client.Client
	ClusterID string
}

func (r *Allocator) IsUpdated(ctx context.Context, _ logr.Logger, object *connectorsv1.StaticAccessKey) (bool, error) {
	res, err := sakeyutils.GetStaticAccessKey(
		ctx, object.Status.KeyID, object.Spec.ServiceAccountID, r.ClusterID, object.Name, r.Sdk,
	)
	if err != nil {
		return false, err
	}
	// We do not need to check secret for existence
	// because there's no way to recreate it
	return res != nil, nil
}

func (r *Allocator) Update(ctx context.Context, log logr.Logger, object *connectorsv1.StaticAccessKey) error {
	res, err := r.Sdk.Create(
		ctx, object.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(r.ClusterID, object.Name),
	)
	if err != nil {
		return fmt.Errorf("error while creating resource: %v", err)
	}

	// Now we need to create a secret with the key
	if err = secret.Put(
		ctx, r.Client, &object.ObjectMeta, sakeyconfig.ShortName, map[string]string{
			"key":    res.AccessKey.KeyId,
			"secret": res.Secret,
		},
	); err != nil {
		// This exact error is a disaster - we have created key, but
		// have not provided secret with secret key and therefore
		// we will inevitably lose it.
		// TODO (covariance) maybe put log.Fatal here?
		return err
	}

	// And we need to update status
	object.Status.SecretName = secret.Name(&object.ObjectMeta, sakeyconfig.ShortName)
	if err = r.Client.Update(ctx, object); err != nil {
		return fmt.Errorf("error while creating resource: %v", err)
	}

	log.Info("resource allocated successfully")
	return nil
}

func (r *Allocator) Cleanup(ctx context.Context, log logr.Logger, object *connectorsv1.StaticAccessKey) error {
	if err := secret.Remove(ctx, r.Client, &object.ObjectMeta, sakeyconfig.ShortName); err != nil {
		return err
	}

	res, err := sakeyutils.GetStaticAccessKey(
		ctx, object.Status.KeyID, object.Spec.ServiceAccountID, r.ClusterID, object.Name, r.Sdk,
	)
	if err != nil {
		return err
	}
	if res == nil {
		return nil
	}

	if err = r.Sdk.Delete(ctx, res.Id); err != nil {
		return err
	}

	log.Info("resource deleted successfully")
	return nil
}
