// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controllers

// TODO (covariance) push events to get via (kubectl get events)
// TODO (covariance) generalize reconciler

import (
	"context"
	"fmt"
	ycrconfig "k8s-connectors/connectors/ycr/pkg"
	ycrutils "k8s-connectors/connectors/ycr/pkg/utils"
	config "k8s-connectors/pkg"
	"k8s-connectors/pkg/configmaps"
	utils "k8s-connectors/pkg/utils"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
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

	// List of initializers that are to be invoked on this object
	// IsUpdated blocks Update, and order of initializers matters,
	// thus if one of initializers fails, subsequent won't be processed.
	updaters := [...]phases.YandexContainerRegistryUpdater{
		// Allocate corresponding resource in cloud (is not blocked by anything)
		&phases.Allocator{
			Sdk: r.sdk,
		},
		// Register finalizer for the object (is blocked by allocation)
		&phases.FinalizerRegistrar{
			Client: &r.Client,
		},
		// In case spec was updated and our cloud registry does not match with
		// spec, we need to update cloud registry (is blocked by allocation)
		&phases.SpecMatcher{
			Sdk: r.sdk,
		},
		// Update status of the object (is blocked by everything mutating)
		&phases.StatusUpdater{
			Sdk:    r.sdk,
			Client: &r.Client,
		},
		// Entrypoint for resource update (is blocked by status update)
		&phases.EntrypointProvider{
			Client: &r.Client,
		},
	}

	// Update all fragments of object, keeping track of whether
	// all of them are initialized
	for _, updater := range updaters {
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
	return !registry.DeletionTimestamp.IsZero() && utils.ContainsString(registry.Finalizers, ycrconfig.RegistryFinalizerName), nil
}

func (r *yandexContainerRegistryReconciler) finalize(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	ycr, err := ycrutils.GetRegistry(ctx, registry, r.sdk)
	if err != nil {
		return err
	}

	if ycr != nil {
		op, err := r.sdk.WrapOperation(r.sdk.ContainerRegistry().Registry().Delete(ctx, &containerregistry.DeleteRegistryRequest{
			RegistryId: ycr.Id,
		}))
		if err != nil {
			// Not found error is already handled by getRegistryId
			return fmt.Errorf("error while deleting registry: %v", err)
		}
		if err := op.Wait(ctx); err != nil {
			return fmt.Errorf("error while deleting registry: %v", err)
		}
	} else {
		// It is assumed that id is the actual id of the object since
		// its lifecycle must be fully managed by connector.
		// id being empty means that it was deleted externally,
		// thus finalization is considered complete.
		log.Info("corresponding object was deleted externally")
	}

	// Also we must remove configmap created as endpoint
	if err := configmaps.Remove(ctx, r.Client, registry); err != nil {
		return fmt.Errorf("unable to remove entrypoint: %v", err)
	}

	// Now we need to state that finalization of this object is no longer needed.
	registry.Finalizers = utils.RemoveString(registry.Finalizers, ycrconfig.RegistryFinalizerName)
	if err := r.Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to remove finalizer: %v", err)
	}

	log.Info("registry successfully deleted")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *yandexContainerRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&connectorsv1.YandexContainerRegistry{}).
		Complete(r)
}
