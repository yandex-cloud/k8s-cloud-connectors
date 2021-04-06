// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controllers

// TODO (covariance) push events to get via (kubectl get events)
// TODO (covariance) generalize reconciler

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
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
	IsUpdated func(context.Context, *connectorsv1.YandexContainerRegistry) (bool, error)
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
		// In case spec was updated and our cloud registry does not match with
		// spec, we need to update cloud registry (is blocked by allocation)
		{
			IsUpdated: r.specMatched,
			Update: r.specMatch,
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
		isInitialized, err := updater.IsUpdated(ctx, &registry)
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

func checkRegistryMatchWithYcr(ycr *containerregistry.Registry, registry *connectorsv1.YandexContainerRegistry) bool {
	cluster, ok1 := ycr.Labels[RegistryCloudClusterLabel]
	name, ok2 := ycr.Labels[RegistryCloudNameLabel]
	return ok1 && ok2 && cluster == registry.ClusterName && name == registry.Name
}

// getRegistryId: tries to retrieve YC ID of registry and check whether it exists
// If registry does not exist, this method returns nil
func (r yandexContainerRegistryReconciler) getRegistry(ctx context.Context, registry *connectorsv1.YandexContainerRegistry) (*containerregistry.Registry, error) {
	// If id is written in the status, we need to check
	// whether it exists in the cloud
	if registry.Status.Id != "" {
		ycr, err := r.sdk.ContainerRegistry().Registry().Get(ctx, &containerregistry.GetRegistryRequest{
			RegistryId: registry.Status.Id,
		})
		if err != nil {
			// If registry was not found then it does not exist,
			// but this error is not fatal, just a mismatch between
			// out status and real world state.
			if errors.CheckRPCErrorNotFound(err) {
				return nil, nil
			}
			// Otherwise, it is fatal
			return nil, fmt.Errorf("cannot get registry from cloud: %v", err)
		}

		// If labels do match with our object, then we have found it
		if checkRegistryMatchWithYcr(ycr, registry) {
			return ycr, nil
		}

		// Otherwise registry is not found, but that is ok:
		// we will try to list resources and find the one we need.
	}

	// TODO (covariance) pagination
	// Otherwise, we try to match cluster name and meta name
	// with registries in the cloud
	list, err := r.sdk.ContainerRegistry().Registry().List(ctx, &containerregistry.ListRegistriesRequest{
		FolderId: registry.Spec.FolderId,
	})
	if err != nil {
		// This error is fatal
		return nil, fmt.Errorf("cannot list registries in folder: %v", err)
	}

	for _, ycr := range list.Registries {
		// If labels do match with our object, then we have found it
		if checkRegistryMatchWithYcr(ycr, registry) {
			return ycr, nil
		}
	}

	// Nothing found, no such registry
	return nil, nil
}

func (r *yandexContainerRegistryReconciler) mustBeFinalized(registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	return !registry.DeletionTimestamp.IsZero() && utils.ContainsString(registry.Finalizers, RegistryFinalizerName), nil
}

func (r *yandexContainerRegistryReconciler) finalize(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	ycr, err := r.getRegistry(ctx, registry)
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

	// Now we need to state that finalization of this object is no longer needed.
	registry.Finalizers = utils.RemoveString(registry.Finalizers, RegistryFinalizerName)
	if err := r.Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to remove finalizer: %v", err)
	}
	log.Info("registry successfully deleted")
	return nil
}

func (r *yandexContainerRegistryReconciler) registryAllocated(ctx context.Context, registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	ycr, err := r.getRegistry(ctx, registry)
	if err != nil {
		return false, fmt.Errorf("unable to get registry: %v", err)
	}

	return ycr != nil, nil
}

func (r *yandexContainerRegistryReconciler) registryAllocate(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
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

	log.Info("registry allocated in cloud")
	return nil
}

func (r *yandexContainerRegistryReconciler) finalizationRegistered(_ context.Context, registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	return utils.ContainsString(registry.Finalizers, RegistryFinalizerName), nil
}

func (r *yandexContainerRegistryReconciler) finalizationRegister(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	registry.Finalizers = append(registry.Finalizers, RegistryFinalizerName)
	if err := r.Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to update registry status: %v", err)
	}
	log.Info("finalizer registered")
	return nil
}

func (r *yandexContainerRegistryReconciler) specMatched(ctx context.Context, registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	ycr, err := r.getRegistry(ctx, registry)
	if err != nil {
		return false, fmt.Errorf("unable to get registry: %v", err)
	}
	if ycr == nil {
		return false, fmt.Errorf("registry %s not found in folder %s", registry.Spec.Name, registry.Spec.FolderId)
	}

	// Here we will check immutable fields
	if registry.Spec.FolderId != "" && ycr.FolderId != registry.Spec.FolderId {
		return false, fmt.Errorf("FolderId changed, invalid state for registry")
	}
	return ycr.Name == registry.Spec.Name, nil
}

func (r *yandexContainerRegistryReconciler) specMatch(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	ycr, err := r.getRegistry(ctx, registry)
	if err != nil {
		return fmt.Errorf("unable to get registry: %v", err)
	}
	if ycr == nil {
		return fmt.Errorf("registry %s not found in folder %s", registry.Spec.Name, registry.Spec.FolderId)
	}

	op, err := r.sdk.WrapOperation(r.sdk.ContainerRegistry().Registry().Update(ctx, &containerregistry.UpdateRegistryRequest{
		RegistryId: ycr.Id,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"name"},
		},
		Name:       registry.Spec.Name,
	}))

	if err != nil {
		return fmt.Errorf("can't update registry in cloud: %v", err)
	}
	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("can't update registry in cloud: %v", err)
	}
	if _, err := op.Response(); err != nil {
		return fmt.Errorf("can't update registry in cloud: %v", err)
	}

	log.Info("registry spec matched with cloud")
	return nil
}

func (r *yandexContainerRegistryReconciler) statusUpdated(_ context.Context, _ *connectorsv1.YandexContainerRegistry) (bool, error) {
	// In every reconciliation we need to update
	// status. Therefore, status is never marked
	// as updated.
	return false, nil
}

func (r *yandexContainerRegistryReconciler) statusUpdate(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	ycr, err := r.getRegistry(ctx, registry)
	if err != nil {
		return fmt.Errorf("unable to get registry: %v", err)
	}
	if ycr == nil {
		return fmt.Errorf("registry %s not found in folder %s", registry.Spec.Name, registry.Spec.FolderId)
	}

	// No type check here is needed, if we cannot find it,
	// it must be something internal, otherwise getRegistryId
	// must already return error

	registry.Status.Id = ycr.Id
	// TODO (covariance) decide what to do with registry.Status.Status
	// TODO (covariance) maybe store registry.Status.CreatedAt as a timestamp?
	registry.Status.CreatedAt = ycr.CreatedAt.String()
	registry.Status.Labels = ycr.Labels

	if err := r.Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to update registry status: %v", err)
	}

	log.Info("registry status updated")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *yandexContainerRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&connectorsv1.YandexContainerRegistry{}).
		Complete(r)
}
