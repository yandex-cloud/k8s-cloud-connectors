// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controllers

// TODO (covariance) push events to get via (kubectl get events)
// TODO (covariance) generalize reconciler

import (
	"context"
	"fmt"
	"k8s-connectors/pkg/config"
	"k8s-connectors/pkg/errors"
	"k8s-connectors/pkg/utils"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"github.com/yandex-cloud/go-sdk"

	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
)

const (
	RegistryCloudClusterLabel string = "managed-kubernetes-cluster-id"
	RegistryCloudNameLabel    string = "managed-kubernetes-registry-metadata-name"
	NoRegistryFound           string = "not-found"
	RegistryFinalizerName     string = "finalizer.yc-registry.connectors.cloud.yandex.ru"
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

type yandexContainerRegistryUpdater struct {
	IsUpdated func(context.Context, logr.Logger, *connectorsv1.YandexContainerRegistry) (bool, error)
	Update    func(context.Context, logr.Logger, *connectorsv1.YandexContainerRegistry) error
}

//+kubebuilder:rbac:groups=connectors.cloud.yandex.ru,resources=yandexcontainerregistries,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=connectors.cloud.yandex.ru,resources=yandexcontainerregistries/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=connectors.cloud.yandex.ru,resources=yandexcontainerregistries/finalizers,verbs=update
func (r *yandexContainerRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithValues("yandexcontainerregistry", req.NamespacedName)

	// Try to retrieve resource from k8s
	var registry connectorsv1.YandexContainerRegistry
	if err := r.Get(ctx, req.NamespacedName, &registry); err != nil {
		// It still can be OK if we have not found it, and we do not need to reconcile it again

		// This outcome signifies that we just cannot find resource, that is ok
		if apierrors.IsNotFound(err) {
			log.Info("Resource not found in k8s, reconciliation not possible")
			return config.GetNormalResult()
		}

		// Some unexpected error occurred, must throw
		return config.GetErroredResult(err)
	}

	// If object must be currently finalized, do it and quit
	mustBeFinalized, err := r.mustBeFinalized(ctx, log, &registry)
	if err != nil {
		return config.GetErroredResult(err)
	}
	if mustBeFinalized {
		if err := r.finalize(ctx, log, &registry); err != nil {
			return config.GetErroredResult(err)
		}
		return config.GetNormalResult()
	}

	// List of initializers that are to be invoked on this object
	// IsUpdated blocks Update, and order of initializers matters,
	// thus if one of initializers fails, subsequent won't be processed.
	updaters := [...]yandexContainerRegistryUpdater{
		// Allocate corresponding resource in cloud (is not blocked by anything)
		{
			IsUpdated: r.registryAllocated,
			Update:    r.registryAllocate,
		},
		// Register finalizer for the object (is blocked by allocation)
		{
			IsUpdated: r.finalizationRegistered,
			Update:    r.finalizationRegister,
		},
		// Update status of the object (is blocked by everything, must be last ops)
		{
			IsUpdated: r.statusUpdated,
			Update:    r.statusUpdate,
		},
	}

	// Update all fragments of object, keeping track of whether
	// all of them are initialized
	for _, updater := range updaters {
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

// getRegistryId: tries to retrieve YC ID of registry and check whether it exists
// If registry does not exist, this method returns NoRegistryFound
func (r yandexContainerRegistryReconciler) getRegistryId(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) (string, error) {
	// If id is written in the status, we need to check
	// whether it exists in the cloud
	if registry.Status.Id != "" {
		_, err := r.sdk.ContainerRegistry().Registry().Get(ctx, &containerregistry.GetRegistryRequest{
			RegistryId: registry.Status.Id,
		})

		if err != nil {
			// If registry was not found then it does not exist,
			// but this error is not fatal
			if errors.CheckRPCErrorNotFound(err) {
				return NoRegistryFound, nil
			}
			// Otherwise, it is fatal
			return NoRegistryFound, fmt.Errorf("cannot get registry from cloud: %v", err)
		}

		return registry.Status.Id, nil
	}

	// TODO (covariance) pagination
	// Otherwise, we try to match cluster name and meta name
	// with registries in the cloud
	list, err := r.sdk.ContainerRegistry().Registry().List(ctx, &containerregistry.ListRegistriesRequest{
		FolderId: registry.Spec.FolderId,
	})
	if err != nil {
		// This error is fatal
		return NoRegistryFound, fmt.Errorf("cannot list registries in folder: %v", err)
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
	return NoRegistryFound, nil
}

func (r *yandexContainerRegistryReconciler) mustBeFinalized(_ context.Context, _ logr.Logger, registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	return !registry.DeletionTimestamp.IsZero() && utils.ContainsString(registry.Finalizers, RegistryFinalizerName), nil
}

func (r *yandexContainerRegistryReconciler) finalize(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	log.Info("deleting registry")
	id, err := r.getRegistryId(ctx, log, registry)
	if err != nil {
		return err
	}

	if id != NoRegistryFound {
		op, err := r.sdk.WrapOperation(r.sdk.ContainerRegistry().Registry().Delete(ctx, &containerregistry.DeleteRegistryRequest{
			RegistryId: id,
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

	// Now we need to state that finalization of this object is no longer needed.
	registry.Finalizers = utils.RemoveString(registry.Finalizers, RegistryFinalizerName)
	if err := r.Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to remove finalizer: %v", err)
	}
	log.Info("registry successfully deleted")
	return nil
}

func (r *yandexContainerRegistryReconciler) registryAllocated(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	id, err := r.getRegistryId(ctx, log, registry)
	if err != nil {
		return false, err
	}

	return id != NoRegistryFound, nil
}

func (r *yandexContainerRegistryReconciler) registryAllocate(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	log.Info("creating registry")
	op, err := r.sdk.WrapOperation(r.sdk.ContainerRegistry().Registry().Create(ctx, &containerregistry.CreateRegistryRequest{
		FolderId: registry.Spec.FolderId,
		Name:     registry.Spec.Name,
		Labels: map[string]string{
			RegistryCloudClusterLabel: registry.ClusterName,
			RegistryCloudNameLabel:    registry.Name,
		},
	}))

	if err != nil {
		// This case is quite strange, but we cannot do anything about it,
		// so we just ignore it.
		if errors.CheckRPCErrorAlreadyExists(err) {
			// TODO (covariance) is it considered error or not?
			log.Info("resource already exists")
			return nil
		}
		return fmt.Errorf("error while creating registry: %v", err)
	}

	if err := op.Wait(ctx); err != nil {
		// According to SDK architecture, we do not actually need
		// to type check here. Every error here is really fatal.
		return fmt.Errorf("error while creating registry: %v", err)
	}

	if _, err := op.Response(); err != nil {
		// If we cannot get response from operation,
		// then it's totally not our responsibility.
		// And, by the way, fatal.
		return fmt.Errorf("error while creating registry: %v", err)
	}

	return nil
}

func (r *yandexContainerRegistryReconciler) finalizationRegistered(_ context.Context, _ logr.Logger, registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	return utils.ContainsString(registry.Finalizers, RegistryFinalizerName), nil
}

func (r *yandexContainerRegistryReconciler) finalizationRegister(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	log.Info("registering finalizer")
	registry.Finalizers = append(registry.Finalizers, RegistryFinalizerName)
	if err := r.Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to update registry status: %v", err)
	}
	log.Info("finalizer registered")
	return nil
}

func (r *yandexContainerRegistryReconciler) statusUpdated(_ context.Context, _ logr.Logger, _ *connectorsv1.YandexContainerRegistry) (bool, error) {
	// In every reconciliation we need to update
	// status. Therefore, status is never marked
	// as updated.
	return false, nil
}

func (r *yandexContainerRegistryReconciler) statusUpdate(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	log.Info("updating object status")
	id, err := r.getRegistryId(ctx, log, registry)
	if err != nil {
		return fmt.Errorf("unable to get registry id: %v", err)
	}
	if id == NoRegistryFound {
		return fmt.Errorf("registry %s not found in folder %s", registry.Spec.Name, registry.Spec.FolderId)
	}

	// No type check here is needed, if we cannot find it,
	// it must be something internal, otherwise getRegistryId
	// must already return error
	op, err := r.sdk.ContainerRegistry().Registry().Get(ctx, &containerregistry.GetRegistryRequest{
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
		return fmt.Errorf("unable to update registry status: %v", err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *yandexContainerRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&connectorsv1.YandexContainerRegistry{}).
		Complete(r)
}
