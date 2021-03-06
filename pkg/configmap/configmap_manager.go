// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package configmap

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rtcl "sigs.k8s.io/controller-runtime/pkg/client"
)

func cmapName(resourceName, kind string) string {
	return kind + "-" + resourceName + "-" + "configmap"
}

func Exists(ctx context.Context, cl rtcl.Client, objectName, namespace, kind string) (bool, error) {
	cmapName := cmapName(objectName, kind)

	var cmapObj v1.ConfigMap
	err := cl.Get(ctx, rtcl.ObjectKey{Namespace: namespace, Name: cmapName}, &cmapObj)
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func Put(ctx context.Context, cl rtcl.Client, objectName, namespace, kind string, data map[string]string) error {
	cmapName := cmapName(objectName, kind)

	var cmapObj v1.ConfigMap
	err := cl.Get(ctx, rtcl.ObjectKey{Namespace: namespace, Name: cmapName}, &cmapObj)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	newState := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmapName,
			Namespace: namespace,
			Labels: map[string]string{
				"kind": kind,
			},
		},
		Data: data,
	}
	if errors.IsNotFound(err) {
		if err := cl.Create(ctx, &newState); err != nil {
			return fmt.Errorf("cannot create configmap: %w", err)
		}
	} else {
		if err := cl.Update(ctx, &newState); err != nil {
			return fmt.Errorf("cannot update configmap: %w", err)
		}
	}
	return nil
}

func Remove(ctx context.Context, cl rtcl.Client, objectName, namespace, kind string) error {
	cmapName := cmapName(objectName, kind)

	var cmapObj v1.ConfigMap
	err := cl.Get(ctx, rtcl.ObjectKey{Namespace: namespace, Name: cmapName}, &cmapObj)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("cannot get configmap: %w", err)
	}

	if err := cl.Delete(ctx, &cmapObj); err != nil {
		return fmt.Errorf("cannot delete configmap: %w", err)
	}

	return nil
}
