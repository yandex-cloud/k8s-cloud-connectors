// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "k8s-connectors/connector/yos/api/v1"
	"k8s-connectors/connector/yos/controller/adapter"
	yosconfig "k8s-connectors/connector/yos/pkg/config"
	"k8s-connectors/pkg/config"
	"k8s-connectors/pkg/util"
)

// yandexObjectStorageReconciler reconciles a YandexContainerRegistry object
type yandexObjectStorageReconciler struct {
	client.Client
	adapter adapter.YandexObjectStorageAdapter
	log     logr.Logger
}

func NewYandexObjectStorageReconciler(
	cl client.Client, log logr.Logger,
) (*yandexObjectStorageReconciler, error) {
	impl, err := adapter.NewYandexObjectStorageAdapterSDK()
	if err != nil {
		return nil, err
	}
	return &yandexObjectStorageReconciler{
		Client:  cl,
		adapter: impl,
		log:     log,
	}, nil
}

// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=yandexobjectstorages,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=yandexobjectstorages/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=yandexobjectstorages/finalizers,verbs=update
// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=staticaccesskeys,verbs=get
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get

func (r *yandexObjectStorageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithValues("name", req.NamespacedName)
	log.V(1).Info("started reconciliation")

	// Try to retrieve object from k8s
	var object connectorsv1.YandexObjectStorage
	if err := r.Get(ctx, req.NamespacedName, &object); err != nil {
		// It still can be OK if we have not found it, and we do not need to reconcile it again

		// This outcome signifies that we just cannot find object, that is ok
		if apierrors.IsNotFound(err) {
			log.V(1).Info("object not found in k8s, reconciliation not possible")
			return config.GetNeverResult()
		}

		// Some unexpected error occurred, must throw
		return config.GetErroredResult(err)
	}

	// If object must be currently finalized, do it and quit
	mustBeFinalized, err := r.mustBeFinalized(&object)
	if err != nil {
		return config.GetErroredResult(err)
	}
	if mustBeFinalized {
		if err := r.finalize(ctx, log, &object); err != nil {
			return config.GetErroredResult(err)
		}
		return config.GetNormalResult()
	}

	if err := util.RegisterFinalizer(
		ctx, r.Client, log, &object.ObjectMeta, &object, yosconfig.FinalizerName,
	); err != nil {
		return config.GetErroredResult(err)
	}

	if err := r.allocateResource(ctx, log, &object); err != nil {
		return config.GetErroredResult(err)
	}

	log.V(1).Info("finished reconciliation")
	return config.GetNormalResult()
}

func (r *yandexObjectStorageReconciler) mustBeFinalized(object *connectorsv1.YandexObjectStorage) (bool, error) {
	return !object.DeletionTimestamp.IsZero() && util.ContainsString(
		object.Finalizers, yosconfig.FinalizerName,
	), nil
}

func (r *yandexObjectStorageReconciler) finalize(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexObjectStorage,
) error {
	finalizationLog := log.WithName("finalization")
	finalizationLog.V(1).Info("started")

	if err := r.deallocateResource(ctx, finalizationLog, object); err != nil {
		return err
	}

	if err := util.RegisterFinalizer(
		ctx, r.Client, finalizationLog, &object.ObjectMeta, object, yosconfig.FinalizerName,
	); err != nil {
		return err
	}

	finalizationLog.Info("object finalized successfully")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *yandexObjectStorageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&connectorsv1.YandexObjectStorage{}).
		Complete(r)
}
