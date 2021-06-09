// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"

	connectorsv1 "k8s-connectors/connector/sakey/api/v1"
	sakeyconfig "k8s-connectors/connector/sakey/pkg/config"
	sakeyutils "k8s-connectors/connector/sakey/pkg/util"
	"k8s-connectors/pkg/secret"
)

func (r *staticAccessKeyReconciler) allocateResource(
	ctx context.Context, log logr.Logger, object *connectorsv1.StaticAccessKey,
) (*awscompatibility.AccessKey, error) {
	log.V(1).Info("started")

	res, err := sakeyutils.GetStaticAccessKey(
		ctx, object.Status.KeyID, object.Spec.ServiceAccountID, r.clusterID, object.Name, r.adapter,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get resource: %v", err)
	}
	if res != nil {
		return res, nil
	}
	response, err := r.adapter.Create(
		ctx, object.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(r.clusterID, object.Name),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create resource: %v", err)
	}

	// Now we need to create a secret with the key
	if err := secret.Put(
		ctx, r.Client, &object.ObjectMeta, sakeyconfig.ShortName, map[string]string{
			"key":    response.AccessKey.KeyId,
			"secret": response.Secret,
		},
	); err != nil {
		// This exact error is a disaster - we have created key, but
		// have not provided secret with secret key and therefore
		// we will inevitably lose it.
		// TODO (covariance) we either need to make some retries here
		//  or delete the key itself because it is useless and possibly dangerous.
		return nil, fmt.Errorf("unable to create secret, this is fatal: %v", err)
	}

	// And we need to update status
	object.Status.SecretName = secret.Name(&object.ObjectMeta, sakeyconfig.ShortName)
	if err := r.Client.Update(ctx, object); err != nil {
		return nil, fmt.Errorf("unable to update object status: %v", err)
	}

	log.Info("successful")
	return response.AccessKey, nil
}

func (r *staticAccessKeyReconciler) deallocateResource(
	ctx context.Context, log logr.Logger, object *connectorsv1.StaticAccessKey,
) error {
	log.V(1).Info("started")

	if err := secret.Remove(ctx, r.Client, &object.ObjectMeta, sakeyconfig.ShortName); err != nil {
		return fmt.Errorf("unable to delete secret: %v", err)
	}

	res, err := sakeyutils.GetStaticAccessKey(
		ctx, object.Status.KeyID, object.Spec.ServiceAccountID, r.clusterID, object.Name, r.adapter,
	)
	if err != nil {
		return fmt.Errorf("unable to get resource: %v", err)
	}
	if res == nil {
		return nil
	}

	if err := r.adapter.Delete(ctx, res.Id); err != nil {
		return fmt.Errorf("unable to delete resource: %v", err)
	}

	log.Info("successful")
	return nil
}
