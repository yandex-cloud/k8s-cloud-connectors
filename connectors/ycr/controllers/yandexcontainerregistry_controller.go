// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controllers

// TODO (covariance) push events to get via (kubectl get events)
// TODO (covariance) generalize reconciler

import (
	"context"
	"fmt"
	ycrconfig "k8s-connectors/connectors/ycr/pkg/config"
	"k8s-connectors/pkg/config"
	"k8s-connectors/pkg/utils"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/yandex-cloud/go-sdk"

	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	"k8s-connectors/connectors/ycr/controllers/phases"
)

// yandexContainerRegistryReconciler reconciles a YandexContainerRegistry object
type yandexContainerRegistryReconciler struct {
	client.Client
	log    logr.Logger
	scheme *runtime.Scheme
	sdk    *ycsdk.SDK

	// phases that are to be invoked on this object
	// IsUpdated blocks Update, and order of initializers matters,
	// thus if one of initializers fails, subsequent won't be processed.
	// Upon destruction of object, phase cleanups are called in
	// reverse order.
	phases []phases.YandexContainerRegistryUpdater
}

func NewYandexContainerRegistryReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) (*yandexContainerRegistryReconciler, error) {
	sdk, err := ycsdk.Build(context.Background(), ycsdk.Config{
		Credentials: ycsdk.InstanceServiceAccount(),
	})
	if err != nil {
		return nil, err
	}
	return &yandexContainerRegistryReconciler{
		Client: client,
		log:    log,
		scheme: scheme,
		sdk:    sdk,
		phases: []phases.YandexContainerRegistryUpdater{
			// Register finalizer for the object (is blocked by allocation)
			&phases.FinalizerRegistrar{
				Client: &client,
			},
			// Allocate corresponding resource in cloud
			// (is blocked by finalizer registration,
			// because otherwise resource can leak)
			&phases.Allocator{
				Sdk: sdk,
			},
			// In case spec was updated and our cloud registry does not match with
			// spec, we need to update cloud registry (is blocked by allocation)
			&phases.SpecMatcher{
				Sdk: sdk,
			},
			// Update status of the object (is blocked by everything mutating)
			&phases.StatusUpdater{
				Sdk:    sdk,
				Client: &client,
			},
			// Entrypoint for resource update (is blocked by status update)
			&phases.EndpointProvider{
				Client: &client,
			},
		},
	}, nil
}

//+kubebuilder:rbac:groups=connectors.cloud.yandex.ru,resources=yandexcontainerregistries,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=connectors.cloud.yandex.ru,resources=yandexcontainerregistries/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=connectors.cloud.yandex.ru,resources=yandexcontainerregistries/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources:configmaps,verbs=create,update,delete
func (r *yandexContainerRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithValues("yandexcontainerregistry", req.NamespacedName)

	// Try to retrieve resource from k8s
	var registry connectorsv1.YandexContainerRegistry
	if err := r.Get(ctx, req.NamespacedName, &registry); err != nil {
		// It still can be OK if we have not found it, and we do not need to reconcile it again

		// This outcome signifies that we just cannot find resource, that is ok
		if apierrors.IsNotFound(err) {
			log.Info("Resource not found in k8s, reconciliation not possible")
			return ctrl.Result{
				RequeueAfter: config.ErroredTimeout,
			}, nil
		}

		// Some unexpected error occurred, must throw
		return ctrl.Result{
			RequeueAfter: config.ErroredTimeout,
		}, err
	}

	// If object must be currently finalized, do it and quit
	mustBeFinalized, err := r.mustBeFinalized(&registry)
	if err != nil {
		return ctrl.Result{
			RequeueAfter: config.ErroredTimeout,
		}, err
	}
	if mustBeFinalized {
		if err := r.finalize(ctx, log, &registry); err != nil {
			return ctrl.Result{
				RequeueAfter: config.ErroredTimeout,
			}, err
		}
		return ctrl.Result{
			RequeueAfter: config.NormalTimeout,
		}, nil
	}

	// Update all fragments of object, keeping track of whether
	// all of them are initialized
	for _, updater := range r.phases {
		isInitialized, err := updater.IsUpdated(ctx, &registry)
		if err != nil {
			return ctrl.Result{
				RequeueAfter: config.ErroredTimeout,
			}, err
		}
		if !isInitialized {
			if err := updater.Update(ctx, log, &registry); err != nil {
				return ctrl.Result{
					RequeueAfter: config.ErroredTimeout,
				}, err
			}
		}
	}

	return ctrl.Result{
		RequeueAfter: config.NormalTimeout,
	}, nil
}

func (r *yandexContainerRegistryReconciler) mustBeFinalized(registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	return !registry.DeletionTimestamp.IsZero() && utils.ContainsString(registry.Finalizers, ycrconfig.FinalizerName), nil
}

func (r *yandexContainerRegistryReconciler) finalize(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	for i := len(r.phases); i != 0; i-- {
		if err := r.phases[i - 1].Cleanup(ctx, log, registry); err != nil {
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
