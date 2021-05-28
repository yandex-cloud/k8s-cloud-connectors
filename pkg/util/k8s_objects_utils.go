// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package util

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NamespacedName(obj client.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

func RegisterFinalizer(
	ctx context.Context, cl client.Client, log logr.Logger, meta *metav1.ObjectMeta, object client.Object,
	finalizer string,
) error {
	if ContainsString(meta.Finalizers, finalizer) {
		return nil
	}
	meta.Finalizers = append(meta.Finalizers, finalizer)
	if err := cl.Update(ctx, object); err != nil {
		return fmt.Errorf("unable to update status: %v", err)
	}
	log.Info("finalizer registered successfully")
	return nil
}

func DeregisterFinalizer(
	ctx context.Context, cl client.Client, log logr.Logger, meta *metav1.ObjectMeta, object client.Object,
	finalizer string,
) error {
	meta.Finalizers = RemoveString(meta.Finalizers, finalizer)
	if err := cl.Update(ctx, object); err != nil {
		return fmt.Errorf("unable to remove finalizer: %v", err)
	}

	log.Info("finalizer removed successfully")
	return nil
}
