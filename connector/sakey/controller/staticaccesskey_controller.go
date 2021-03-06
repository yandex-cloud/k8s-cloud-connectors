// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	ycsdk "github.com/yandex-cloud/go-sdk"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/api/v1"
	"github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/controller/adapter"
	sakeyconfig "github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/pkg/config"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/config"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/phase"
)

// staticAccessKeyReconciler reconciles a StaticAccessKey object
type staticAccessKeyReconciler struct {
	client.Client
	adapter   adapter.StaticAccessKeyAdapter
	log       logr.Logger
	clusterID string
}

func NewStaticAccessKeyReconciler(log logr.Logger, cl client.Client,
	sdk *ycsdk.SDK, clusterID string) *staticAccessKeyReconciler {
	return &staticAccessKeyReconciler{
		Client:    cl,
		adapter:   adapter.NewStaticAccessKeyAdapter(sdk),
		log:       log,
		clusterID: clusterID,
	}
}

// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=staticaccesskeys,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=staticaccesskeys/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=staticaccesskeys/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *staticAccessKeyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithValues("name", req.NamespacedName)
	log.V(1).Info("started reconciliation")

	// Try to retrieve object from k8s
	var object connectorsv1.StaticAccessKey
	if err := r.Get(ctx, req.NamespacedName, &object); err != nil {
		// This outcome signifies that we just cannot find object, that is OK,
		// we just never want to reconcile it again unless triggered externally.
		if apierrors.IsNotFound(err) {
			log.V(1).Info("object not found in k8s, reconciliation not possible")
			return config.GetNeverResult()
		}

		return config.GetErroredResult(fmt.Errorf("unable to get object from k8s: %w", err))
	}

	// If object must be currently finalized, do it and quit
	if phase.MustBeFinalized(&object.ObjectMeta, sakeyconfig.FinalizerName) {
		if err := r.finalize(ctx, log.WithName("finalize"), &object); err != nil {
			return config.GetErroredResult(fmt.Errorf("unable to finalize object: %w", err))
		}
		return config.GetNormalResult()
	}

	if err := phase.RegisterFinalizer(
		ctx, r.Client, log.WithName("register-finalizer"), &object.ObjectMeta, &object, sakeyconfig.FinalizerName,
	); err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to register finalizer: %w", err))
	}

	res, err := r.allocateResource(ctx, log.WithName("allocate-resource"), &object)
	if err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to allocate resource: %w", err))
	}

	if err := r.updateStatus(ctx, log.WithName("update-status"), &object, res); err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to update status: %w", err))
	}

	log.V(1).Info("finished reconciliation")
	return config.GetNormalResult()
}

func (r *staticAccessKeyReconciler) finalize(
	ctx context.Context, log logr.Logger, object *connectorsv1.StaticAccessKey,
) error {
	log.V(1).Info("started")

	if err := r.deallocateResource(ctx, log.WithName("deallocate-resource"), object); err != nil {
		return fmt.Errorf("unable to deallocate resource: %w", err)
	}

	if err := phase.DeregisterFinalizer(
		ctx, r.Client, log.WithName("deregister-finalizer"), &object.ObjectMeta, object, sakeyconfig.FinalizerName,
	); err != nil {
		return fmt.Errorf("unable to deregister finalizer: %w", err)
	}

	log.Info("successful")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *staticAccessKeyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&connectorsv1.StaticAccessKey{}).
		Complete(r)
}
