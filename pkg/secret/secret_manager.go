// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package secret

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rtcl "sigs.k8s.io/controller-runtime/pkg/client"
)

func Name(objectName, kind string) string {
	return kind + "-" + objectName + "-" + "secret"
}

func Exists(ctx context.Context, client rtcl.Client, objectName, namespace, kind string) (bool, error) {
	secretName := Name(objectName, kind)

	var secretObj v1.Secret
	err := client.Get(ctx, rtcl.ObjectKey{Namespace: namespace, Name: secretName}, &secretObj)
	if errors.IsNotFound(err) {
		return false, nil
	}
	return true, fmt.Errorf("cannot get secret: %w", err)
}

func Put(
	ctx context.Context, client rtcl.Client, objectName, namespace, kind string, data map[string]string,
) error {
	secretName := Name(objectName, kind)

	var secretObj v1.Secret
	err := client.Get(ctx, rtcl.ObjectKey{Namespace: namespace, Name: secretName}, &secretObj)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	newState := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"kind": kind,
			},
		},
		StringData: data,
	}

	if errors.IsNotFound(err) {
		if err = client.Create(ctx, &newState); err != nil {
			return fmt.Errorf("cannot create secret: %w", err)
		}
	} else {
		if err = client.Update(ctx, &newState); err != nil {
			return fmt.Errorf("cannot update secret: %w", err)
		}
	}
	return nil
}

func Remove(ctx context.Context, client rtcl.Client, objectName, namespace, kind string) error {
	secretName := Name(objectName, kind)

	var secretObj v1.Secret
	err := client.Get(ctx, rtcl.ObjectKey{Namespace: namespace, Name: secretName}, &secretObj)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("cannot get secret: %w", err)
	}

	if errors.IsNotFound(err) {
		return nil
	}

	if err = client.Delete(ctx, &secretObj); err != nil {
		return fmt.Errorf("cannot delete secret: %w", err)
	}

	return nil
}
