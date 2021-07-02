// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "github.com/yandex-cloud/k8s-cloud-connectors/connector/ycr/api/v1"
	"github.com/yandex-cloud/k8s-cloud-connectors/connector/ycr/controller/adapter"
	ycrconfig "github.com/yandex-cloud/k8s-cloud-connectors/connector/ycr/pkg/config"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/config"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/phase"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/util"
)

// yandexContainerRegistryReconciler reconciles a YandexContainerRegistry object
type yandexContainerRegistryReconciler struct {
	client.Client
	adapter   adapter.YandexContainerRegistryAdapter
	log       logr.Logger
	clusterID string
}

func NewYandexContainerRegistryReconciler(
	cl client.Client, log logr.Logger, clusterID string,
) (*yandexContainerRegistryReconciler, error) {
	impl, err := adapter.NewYandexContainerRegistryAdapterSDK()
	if err != nil {
		return nil, err
	}
	return &yandexContainerRegistryReconciler{
		Client:    cl,
		adapter:   impl,
		log:       log,
		clusterID: clusterID,
	}, nil
}

// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=yandexcontainerregistries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=yandexcontainerregistries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=yandexcontainerregistries/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *yandexContainerRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithValues("name", req.NamespacedName)
	log.V(1).Info("started reconciliation")

	// Try to retrieve object from k8s
	var object connectorsv1.YandexContainerRegistry
	if err := r.Get(ctx, req.NamespacedName, &object); err != nil {
		// It still can be OK if we have not found it, and we do not need to reconcile it again

		// This outcome signifies that we just cannot find object, that is ok
		if apierrors.IsNotFound(err) {
			log.V(1).Info("object not found in k8s, reconciliation not possible")
			return config.GetNeverResult()
		}

		return config.GetErroredResult(fmt.Errorf("unable to get object from k8s: %w", err))
	}

	// If object must be currently finalized, do it and quit
	mustBeFinalized, err := r.mustBeFinalized(&object)
	if err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to check if object must be finalized: %w", err))
	}
	if mustBeFinalized {
		if err := r.finalize(ctx, log.WithName("finalize"), &object); err != nil {
			return config.GetErroredResult(fmt.Errorf("unable to finalize object: %w", err))
		}
		return config.GetNormalResult()
	}

	if err := phase.RegisterFinalizer(
		ctx, r.Client, log, &object.ObjectMeta, &object, ycrconfig.FinalizerName,
	); err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to register finalizer: %w", err))
	}

	res, err := r.allocateResource(ctx, log.WithName("allocate-resource"), &object)
	if err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to allocate resource: %w", err))
	}

	if err := r.matchSpec(ctx, log.WithName("match-spec"), &object, res); err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to match spec: %w", err))
	}

	if err := r.updateStatus(ctx, log.WithName("update-status"), &object, res); err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to update status: %w", err))
	}

	if err := phase.ProvideConfigmap(
		ctx,
		r.Client,
		log.WithName("provide-configmap"),
		object.Name, ycrconfig.ShortName, object.Namespace,
		map[string]string{"ID": object.Status.ID},
	); err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to provide configmap: %w", err))
	}

	log.V(1).Info("finished reconciliation")
	return config.GetNormalResult()
}

func (r *yandexContainerRegistryReconciler) mustBeFinalized(object *connectorsv1.YandexContainerRegistry) (
	bool, error,
) {
	return !object.DeletionTimestamp.IsZero() && util.ContainsString(
		object.Finalizers, ycrconfig.FinalizerName,
	), nil
}

func (r *yandexContainerRegistryReconciler) finalize(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry,
) error {
	log.V(1).Info("started")

	if err := phase.RemoveConfigmap(
		ctx,
		r.Client,
		log.WithName("remove-configmap"),
		object.Name, ycrconfig.ShortName, object.Namespace,
	); err != nil {
		return fmt.Errorf("unable to remove configmap: %w", err)
	}

	if err := r.deallocateResource(ctx, log.WithName("deallocate-resource"), object); err != nil {
		return fmt.Errorf("unable to deallocate resource: %w", err)
	}

	if err := phase.DeregisterFinalizer(
		ctx, r.Client, log.WithName("deregister-finalizer"), &object.ObjectMeta, object, ycrconfig.FinalizerName,
	); err != nil {
		return fmt.Errorf("unable to deregister finalizer: %w", err)
	}

	log.Info("successful")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *yandexContainerRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&connectorsv1.YandexContainerRegistry{}).
		Complete(r)
}
