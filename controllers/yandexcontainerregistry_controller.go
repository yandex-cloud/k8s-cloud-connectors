// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"github.com/yandex-cloud/go-sdk"

	connectorsv1 "k8s-connectors/api/v1"
	"k8s-connectors/commons"
)

// TODO (covariance) push events to get via (kubectl get events)
// TODO (covariance) generalize reconciler

// YandexContainerRegistryReconciler reconciles a YandexContainerRegistry object
type YandexContainerRegistryReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type YandexContainerRegistryInitializer struct {
	IsInitialized func(context.Context, *ycsdk.SDK, logr.Logger, *connectorsv1.YandexContainerRegistry) (bool, error)
	Initialize    func(context.Context, *ycsdk.SDK, logr.Logger, *connectorsv1.YandexContainerRegistry) error
}

// +kubebuilder:rbac:groups=connectors.cloud.yandex.ru,resources=yandexcontainerregistries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=connectors.cloud.yandex.ru,resources=yandexcontainerregistries/status,verbs=get;update;patch
func (r *YandexContainerRegistryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("yandexcontainerregistry", req.NamespacedName)

	// Try to retrieve resource from k8s
	var registry connectorsv1.YandexContainerRegistry
	if err := r.Get(ctx, req.NamespacedName, &registry); err != nil {
		// It still can be OK if we have not found it, and we do not need to reconcile it again

		// This outcome signifies that we just cannot find resource, that is ok
		if apierrors.IsNotFound(err) {
			log.Info("Resource not found in k8s, reconciliation not possible")
			return commons.GetNormalResult()
		}

		// Some unexpected error occurred, must throw
		return commons.GetErroredResult(err)
	}

	// Building Yandex Cloud SDK for our requests
	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: ycsdk.InstanceServiceAccount(),
	})
	if err != nil {
		return commons.GetErroredResult(err)
	}

	// If object must be currently finalized, do it and quit
	mustBeFinalized, err := r.mustBeFinalized(ctx, sdk, log, &registry)
	if err != nil {
		return commons.GetErroredResult(err)
	}
	if mustBeFinalized {
		if err := r.finalize(ctx, sdk, log, &registry); err != nil {
			return commons.GetErroredResult(err)
		}
		return commons.GetNormalResult()
	}

	// List of initializers that are to be invoked on this object
	// IsInitialized blocks Initialize, and order of initializers matters,
	// thus if one of initializers fails, subsequent won't be processed.
	initializers := [...]YandexContainerRegistryInitializer{
		// Allocate corresponding resource in cloud (is not blocked by anything)
		{
			IsInitialized: r.registryAllocated,
			Initialize:    r.registryAllocate,
		},
		// Register finalizer for the object (is blocked by allocation)
		{
			IsInitialized: r.finalizationRegistered,
			Initialize:    r.finalizationRegister,
		},
		// Update status of the object (is blocked by everything, must be last ops)
		{
			IsInitialized: r.statusUpdated,
			Initialize:    r.statusUpdate,
		},
	}

	// Initialize all fragments of object, keeping track of whether
	// all of them are initialized
	for _, initializer := range initializers {
		isInitialized, err := initializer.IsInitialized(ctx, sdk, log, &registry)
		if err != nil {
			return commons.GetErroredResult(err)
		}
		if !isInitialized {
			if err := initializer.Initialize(ctx, sdk, log, &registry); err != nil {
				return commons.GetErroredResult(err)
			}
		}
	}

	return commons.GetNormalResult()
}

const RegistryCloudClusterLabel = "cluster-mk8s-connectors-cloud-yandex-ru"
const RegistryCloudNameLabel = "name-mk8s-connectors-cloud-yandex-ru"

// getRegistryId: tries to retrieve YC ID of registry and check whether it exists
// If registry does not exist, this method returns empty string
func (r YandexContainerRegistryReconciler) getRegistryId(ctx context.Context, sdk *ycsdk.SDK, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) (string, error) {
	// If it is written in the status, we need to check
	// whether it exists in the cloud
	if registry.Status.Id != "" {
		_, err := sdk.ContainerRegistry().Registry().Get(ctx, &containerregistry.GetRegistryRequest{
			RegistryId: registry.Status.Id,
		})

		if err != nil {
			// If registry was not found then it does not exist,
			// but this error is not fatal
			if commons.CheckRPCErrorNotFound(err) {
				return "", nil
			}
			// Otherwise, it is fatal
			log.Error(err, "cannot get registry from cloud")
			return "", err
		}

		return registry.Status.Id, nil
	}

	// TODO (covariance) pagination
	// Otherwise, we try to match cluster name and meta name
	// with registries in the cloud
	list, err := sdk.ContainerRegistry().Registry().List(ctx, &containerregistry.ListRegistriesRequest{
		FolderId: registry.Spec.FolderId,
	})
	if err != nil {
		// This error is fatal
		log.Error(err, "cannot list registries in folder")
		return "", err
	}

	for _, el := range list.Registries {
		// If labels do match with our object, then we have found it
		cluster, ok1 := el.Labels[RegistryCloudClusterLabel]
		name, ok2 := el.Labels[RegistryCloudNameLabel]
		if ok1 && ok2 && cluster == registry.ClusterName && name == registry.Name {
			return el.Id, nil
		}
	}

	// Nothing found, no such registry
	return "", nil
}

const RegistryFinalizerName = "finalizer.yc-registry.connectors.cloud.yandex.ru"

func (r *YandexContainerRegistryReconciler) mustBeFinalized(_ context.Context, _ *ycsdk.SDK, _ logr.Logger, registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	return !registry.DeletionTimestamp.IsZero() && commons.ContainsString(registry.Finalizers, RegistryFinalizerName), nil
}

func (r *YandexContainerRegistryReconciler) finalize(ctx context.Context, sdk *ycsdk.SDK, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	log.Info("deleting registry")
	id, err := r.getRegistryId(ctx, sdk, log, registry)
	if err != nil {
		return err
	}

	if id != "" {
		op, err := sdk.WrapOperation(sdk.ContainerRegistry().Registry().Delete(ctx, &containerregistry.DeleteRegistryRequest{
			RegistryId: id,
		}))
		if err != nil {
			// Not found error is already handled by getRegistryId
			log.Error(err, "error while deleting registry")
			return err
		}
		if err := op.Wait(ctx); err != nil {
			log.Error(err, "error while deleting registry")
			return err
		}
	} else {
		// It is assumed that id is the actual id of the object since
		// its lifecycle must be fully managed by connector.
		// id being empty means that it was deleted externally,
		// thus finalization is considered complete.
		log.Info("corresponding object was deleted externally")
	}

	// Now we need to state that finalization of this object is no longer needed.
	registry.Finalizers = commons.RemoveString(registry.Finalizers, RegistryFinalizerName)
	if err := r.Update(ctx, registry); err != nil {
		log.Error(err, "unable to remove finalizer")
		return err
	}
	log.Info("registry successfully deleted")
	return nil
}

func (r *YandexContainerRegistryReconciler) registryAllocated(ctx context.Context, sdk *ycsdk.SDK, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	id, err := r.getRegistryId(ctx, sdk, log, registry)
	if err != nil {
		return false, err
	}

	return id != "", nil
}

func (r *YandexContainerRegistryReconciler) registryAllocate(ctx context.Context, sdk *ycsdk.SDK, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	log.Info("creating registry")
	op, err := sdk.WrapOperation(sdk.ContainerRegistry().Registry().Create(ctx, &containerregistry.CreateRegistryRequest{
		FolderId: registry.Spec.FolderId,
		Name:     registry.Spec.Name,
		Labels: map[string]string{
			RegistryCloudClusterLabel: registry.ClusterName,
			RegistryCloudNameLabel:    registry.Name,
			// TODO (covariance) do we need to push k8s labels to cloud? idk, guess no
		},
	}))

	if err != nil {
		// This case is quite strange, but we cannot do anything about it,
		// so we just ignore it.
		if commons.CheckRPCErrorAlreadyExists(err) {
			log.Info("resource already exists")
			return nil
		}
		log.Error(err, "error while creating registry")
		return err
	}

	if err := op.Wait(ctx); err != nil {
		// According to SDK architecture, we do not actually need
		// to type check here. Every error here is really fatal.
		log.Error(err, "error while creating registry")
		return err
	}

	if _, err := op.Response(); err != nil {
		log.Error(err, "error while creating registry")
		return err
	}

	return nil
}

func (r *YandexContainerRegistryReconciler) finalizationRegistered(_ context.Context, _ *ycsdk.SDK, _ logr.Logger, registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	return commons.ContainsString(registry.Finalizers, RegistryFinalizerName), nil
}

func (r *YandexContainerRegistryReconciler) finalizationRegister(ctx context.Context, _ *ycsdk.SDK, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	log.Info("registering finalizer")
	registry.Finalizers = append(registry.Finalizers, RegistryFinalizerName)
	if err := r.Update(ctx, registry); err != nil {
		log.Error(err, "unable to update registry status")
		return err
	}
	log.Info("finalizer registered")
	return nil
}

func (r *YandexContainerRegistryReconciler) statusUpdated(_ context.Context, _ *ycsdk.SDK, _ logr.Logger, _ *connectorsv1.YandexContainerRegistry) (bool, error) {
	// In every reconciliation we need to update
	// status. Therefore, status is never marked
	// as updated.
	return false, nil
}

func (r *YandexContainerRegistryReconciler) statusUpdate(ctx context.Context, sdk *ycsdk.SDK, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	log.Info("updating object status")
	id, err := r.getRegistryId(ctx, sdk, log, registry)
	if err != nil {
		log.Error(err, "unable to get registry id")
		return err
	}
	if id == "" {
		err := commons.ResourceNotFoundError{
			ResourceId: id,
			FolderId:   registry.Spec.FolderId,
		}
		log.Error(err, "unable to update status")
		return err
	}

	// No type check here is needed, if we cannot find it,
	// it must be something internal, otherwise getRegistryId
	// must already return error
	op, err := sdk.ContainerRegistry().Registry().Get(ctx, &containerregistry.GetRegistryRequest{
		RegistryId: id,
	})
	if err != nil {
		return err
	}

	registry.Status.Id = op.Id
	registry.Status.FolderId = op.FolderId
	registry.Status.Name = op.Name
	// TODO (covariance) decide what to do with registry.Status.Status
	// TODO (covariance) maybe store registry.Status.CreatedAt as a timestamp?
	registry.Status.CreatedAt = op.CreatedAt.String()
	registry.Status.Labels = op.Labels

	if err := r.Update(ctx, registry); err != nil {
		log.Error(err, "unable to update status")
		return err
	}

	return nil
}

func (r *YandexContainerRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&connectorsv1.YandexContainerRegistry{}).
		Complete(r)
}
