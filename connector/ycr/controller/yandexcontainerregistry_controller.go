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

	connectorsv1 "k8s-connectors/connector/ycr/api/v1"
	"k8s-connectors/connector/ycr/controller/adapter"
	"k8s-connectors/connector/ycr/controller/phase"
	ycrconfig "k8s-connectors/connector/ycr/pkg/config"
	"k8s-connectors/pkg/config"
	"k8s-connectors/pkg/util"
)

// yandexContainerRegistryReconciler reconciles a YandexContainerRegistry object
type yandexContainerRegistryReconciler struct {
	client.Client
	log       logr.Logger
	clusterID string

	// phases that are to be invoked on this object
	// IsUpdated blocks Update, and order of initializers matters,
	// thus if one of initializers fails, subsequent won't be processed.
	// Upon destruction of object, phase cleanups are called in
	// reverse order.
	phases []phase.YandexContainerRegistryPhase
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
		log:       log,
		clusterID: clusterID,
		phases: []phase.YandexContainerRegistryPhase{
			// Register finalizer for the object (is blocked by allocation)
			&phase.FinalizerRegistrar{
				Client: cl,
			},
			// Allocate corresponding resource in cloud
			// (is blocked by finalizer registration,
			// because otherwise resource can leak)
			&phase.Allocator{
				Sdk:       impl,
				ClusterID: clusterID,
			},
			// In case spec was updated and our cloud registry does not match with
			// spec, we need to update cloud registry (is blocked by allocation)
			&phase.SpecMatcher{
				Sdk:       impl,
				ClusterID: clusterID,
			},
			// Update status of the object (is blocked by everything mutating)
			&phase.StatusUpdater{
				Sdk:       impl,
				Client:    cl,
				ClusterID: clusterID,
			},
			// Entrypoint for resource update (is blocked by status update)
			&phase.EndpointProvider{
				Client: cl,
			},
		},
	}, nil
}

// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=yandexcontainerregistries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=yandexcontainerregistries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=yandexcontainerregistries/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *yandexContainerRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithValues(ycrconfig.LongName, req.NamespacedName)

	// Try to retrieve resource from k8s
	var registry connectorsv1.YandexContainerRegistry
	if err := r.Get(ctx, req.NamespacedName, &registry); err != nil {
		// It still can be OK if we have not found it, and we do not need to reconcile it again

		// This outcome signifies that we just cannot find resource, that is ok
		if apierrors.IsNotFound(err) {
			log.Info("Resource not found in k8s, reconciliation not possible")
			return config.GetNeverResult()
		}

		// Some unexpected error occurred, must throw
		return config.GetErroredResult(err)
	}

	// If object must be currently finalized, do it and quit
	mustBeFinalized, err := r.mustBeFinalized(&registry)
	if err != nil {
		return config.GetErroredResult(err)
	}
	if mustBeFinalized {
		if err := r.finalize(ctx, log, &registry); err != nil {
			return config.GetErroredResult(err)
		}
		return config.GetNormalResult()
	}

	// Update all fragments of object, keeping track of whether
	// all of them are initialized
	for _, updater := range r.phases {
		isInitialized, err := updater.IsUpdated(ctx, log, &registry)
		if err != nil {
			return config.GetErroredResult(err)
		}
		if !isInitialized {
			if err := updater.Update(ctx, log, &registry); err != nil {
				return config.GetErroredResult(err)
			}
		}
	}

	return config.GetNormalResult()
}

func (r *yandexContainerRegistryReconciler) mustBeFinalized(registry *connectorsv1.YandexContainerRegistry) (
	bool, error,
) {
	return !registry.DeletionTimestamp.IsZero() && util.ContainsString(
		registry.Finalizers, ycrconfig.FinalizerName,
	), nil
}

func (r *yandexContainerRegistryReconciler) finalize(
	ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry,
) error {
	for i := len(r.phases); i != 0; i-- {
		if err := r.phases[i-1].Cleanup(ctx, log, registry); err != nil {
			return fmt.Errorf("error during finalization: %v", err)
		}
	}
	log.Info("registry finalized successfully")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *yandexContainerRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&connectorsv1.YandexContainerRegistry{}).
		Complete(r)
}
