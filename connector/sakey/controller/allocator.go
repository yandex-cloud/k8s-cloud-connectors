// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
	"go.uber.org/multierr"

	connectorsv1 "k8s-connectors/connector/sakey/api/v1"
	sakeyconfig "k8s-connectors/connector/sakey/pkg/config"
	sakeyutils "k8s-connectors/connector/sakey/pkg/util"
	"k8s-connectors/pkg/errorhandling"
	"k8s-connectors/pkg/secret"
)

func (r *staticAccessKeyReconciler) allocateResource(
	ctx context.Context, log logr.Logger, object *connectorsv1.StaticAccessKey,
) (*awscompatibility.AccessKey, error) {
	log.V(1).Info("started")

	res, err := sakeyutils.GetStaticAccessKey(
		ctx, object.Status.KeyID, object.Spec.ServiceAccountID, r.clusterID, object.Name, r.adapter,
	)
	if err == nil {
		return res, nil
	}
	if !errorhandling.CheckConnectorErrorCode(err, sakeyconfig.ErrCodeSAKeyNotFound) {
		return nil, fmt.Errorf("unable to get resource: %w", err)
	}
	response, err := r.adapter.Create(
		ctx, object.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(r.clusterID, object.Name),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create resource: %w", err)
	}

	// Now we need to create a secret with the key
	if err := secret.Put(
		ctx, r.Client, object.Name, object.Namespace, sakeyconfig.ShortName, map[string]string{
			"key":    response.AccessKey.KeyId,
			"secret": response.Secret,
		},
	); err != nil {
		// If we cannot create secret, we will just delete key
		// and try again on the next reconciliation
		err := fmt.Errorf("unable to create secret: %w", err)
		if err2 := r.adapter.Delete(ctx, response.AccessKey.KeyId); err2 != nil {
			return nil, multierr.Append(err, fmt.Errorf("unable to delete SAKey in the cloud: %w", err2))
		}
		return nil, err
	}

	// And we need to update status
	object.Status.SecretName = secret.Name(object.Name, sakeyconfig.ShortName)
	if err := r.Client.Update(ctx, object); err != nil {
		return nil, fmt.Errorf("unable to update object status: %w", err)
	}

	log.Info("successful")
	return response.AccessKey, nil
}

func (r *staticAccessKeyReconciler) deallocateResource(
	ctx context.Context, log logr.Logger, object *connectorsv1.StaticAccessKey,
) error {
	log.V(1).Info("started")

	if err := secret.Remove(ctx, r.Client, object.Name, object.Namespace, sakeyconfig.ShortName); err != nil {
		return fmt.Errorf("unable to delete secret: %w", err)
	}

	res, err := sakeyutils.GetStaticAccessKey(
		ctx, object.Status.KeyID, object.Spec.ServiceAccountID, r.clusterID, object.Name, r.adapter,
	)
	if err != nil {
		if errorhandling.CheckConnectorErrorCode(err, sakeyconfig.ErrCodeSAKeyNotFound) {
			log.Info("already deleted")
			return nil
		}
		return fmt.Errorf("unable to get resource: %w", err)
	}

	if err := r.adapter.Delete(ctx, res.Id); err != nil {
		return fmt.Errorf("unable to delete resource: %w", err)
	}

	log.Info("successful")
	return nil
}
