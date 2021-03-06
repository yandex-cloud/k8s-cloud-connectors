// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "github.com/yandex-cloud/k8s-cloud-connectors/connector/yos/api/v1"
	"github.com/yandex-cloud/k8s-cloud-connectors/connector/yos/controller/adapter"
	yosconfig "github.com/yandex-cloud/k8s-cloud-connectors/connector/yos/pkg/config"
	yosutils "github.com/yandex-cloud/k8s-cloud-connectors/connector/yos/pkg/util"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/awsutils"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/config"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/phase"
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

		return config.GetErroredResult(fmt.Errorf("unable to get object from k8s: %w", err))
	}

	cred, err := awsutils.CredentialsFromStaticAccessKey(ctx, object.Namespace, object.Spec.SAKeyName, r.Client)
	if err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to retrieve credentials: %w", err))
	}
	sdk, err := yosutils.NewS3Client(ctx, cred)
	if err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to build sdk: %w", err))
	}

	// If object must be currently finalized, do it and quit
	if phase.MustBeFinalized(&object.ObjectMeta, yosconfig.FinalizerName) {
		if err := r.finalize(ctx, log.WithName("finalize"), &object, sdk); err != nil {
			return config.GetErroredResult(fmt.Errorf("unable to finalize object: %w", err))
		}
		return config.GetNormalResult()
	}

	if err := phase.RegisterFinalizer(
		ctx, r.Client, log.WithName("register-finalizer"), &object.ObjectMeta, &object, yosconfig.FinalizerName,
	); err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to register finalizer: %w", err))
	}

	if err := r.allocateResource(ctx, log.WithName("allocate-resource"), &object, sdk); err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to allocate resource: %w", err))
	}

	if err := phase.ProvideConfigmap(
		ctx,
		r.Client,
		log.WithName("provide-configmap"),
		object.Name, yosconfig.ShortName, object.Namespace,
		map[string]string{"name": object.Spec.Name},
	); err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to provide configmap: %w", err))
	}

	log.V(1).Info("finished reconciliation")
	return config.GetNormalResult()
}

func (r *yandexObjectStorageReconciler) finalize(
	ctx context.Context,
	log logr.Logger,
	object *connectorsv1.YandexObjectStorage,
	sdk *s3.S3,
) error {
	log.V(1).Info("started")

	if err := phase.RemoveConfigmap(
		ctx,
		r.Client,
		log.WithName("provide-configmap"),
		object.Name, yosconfig.ShortName, object.Namespace,
	); err != nil {
		return fmt.Errorf("unable to remove configmap: %w", err)
	}

	if err := r.deallocateResource(ctx, log.WithName("deallocate-resource"), object, sdk); err != nil {
		return fmt.Errorf("unable to deallocate resource: %w", err)
	}

	if err := phase.DeregisterFinalizer(
		ctx, r.Client, log.WithName("finalizer-deregister"), &object.ObjectMeta, object, yosconfig.FinalizerName,
	); err != nil {
		return fmt.Errorf("unable to deregister finalizer: %w", err)
	}

	log.Info("successful")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *yandexObjectStorageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&connectorsv1.YandexObjectStorage{}).
		Complete(r)
}
