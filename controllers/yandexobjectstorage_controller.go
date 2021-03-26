// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "k8s-connectors/api/v1"
)

// YandexObjectStorageReconciler reconciles a YandexObjectStorage object
type YandexObjectStorageReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=connectors.cloud.yandex.ru,resources=yandexobjectstorages,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=connectors.cloud.yandex.ru,resources=yandexobjectstorages/status,verbs=get;update;patch
func (r *YandexObjectStorageReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("yandexobjectstorage", req.NamespacedName)

	// TODO (covariance) write all necessary logic for s3

	return ctrl.Result{}, nil
}

func (r *YandexObjectStorageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&connectorsv1.YandexObjectStorage{}).
		Complete(r)
}
