// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package secrets

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rtcl "sigs.k8s.io/controller-runtime/pkg/client"
)

func SecretName(object *metav1.ObjectMeta, kind string) string {
	return kind + "-" + object.Name + "-" + "secret"
}

func Exists(ctx context.Context, client rtcl.Client, object *metav1.ObjectMeta, kind string) (bool, error) {
	secretName := SecretName(object, kind)

	var secretObj v1.Secret
	err := client.Get(ctx, rtcl.ObjectKey{Namespace: object.Namespace, Name: secretName}, &secretObj)
	if errors.IsNotFound(err) {
		return false, nil
	}
	return true, fmt.Errorf("cannot get secret: %v", err)
}

func Put(
	ctx context.Context, client rtcl.Client, object *metav1.ObjectMeta, kind string, data map[string]string,
) error {
	secretName := SecretName(object, kind)

	var secretObj v1.Secret
	err := client.Get(ctx, rtcl.ObjectKey{Namespace: object.Namespace, Name: secretName}, &secretObj)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	newState := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: object.Namespace,
			Labels: map[string]string{
				"kind": kind,
			},
		},
		StringData: data,
	}

	if errors.IsNotFound(err) {
		if err = client.Create(ctx, &newState); err != nil {
			return fmt.Errorf("cannot create secret: %v", err)
		}
	} else {
		if err = client.Update(ctx, &newState); err != nil {
			return fmt.Errorf("cannot update secret: %v", err)
		}
	}
	return nil
}

func Remove(ctx context.Context, client rtcl.Client, object *metav1.ObjectMeta, kind string) error {
	secretName := SecretName(object, kind)

	var secretObj v1.Secret
	err := client.Get(ctx, rtcl.ObjectKey{Namespace: object.Namespace, Name: secretName}, &secretObj)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("cannot get secret: %v", err)
	}

	if errors.IsNotFound(err) {
		return nil
	}

	if err = client.Delete(ctx, &secretObj); err != nil {
		return fmt.Errorf("cannot delete secret: %v", err)
	}

	return nil
}
