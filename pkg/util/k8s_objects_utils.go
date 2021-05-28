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
	log.V(1).Info("started")
	if ContainsString(meta.Finalizers, finalizer) {
		return nil
	}
	meta.Finalizers = append(meta.Finalizers, finalizer)
	if err := cl.Update(ctx, object); err != nil {
		return fmt.Errorf("unable to register finalizer: %v", err)
	}
	log.Info("successful")
	return nil
}

func DeregisterFinalizer(
	ctx context.Context, cl client.Client, log logr.Logger, meta *metav1.ObjectMeta, object client.Object,
	finalizer string,
) error {
	log.V(1).Info("started")
	meta.Finalizers = RemoveString(meta.Finalizers, finalizer)
	if err := cl.Update(ctx, object); err != nil {
		return fmt.Errorf("unable to deregister finalizer: %v", err)
	}

	log.Info("successful")
	return nil
}
