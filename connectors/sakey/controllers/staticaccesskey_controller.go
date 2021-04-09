// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controllers

// TODO (covariance) push events to get via (kubectl get events)
// TODO (covariance) generalize reconciler

import (
	"context"
	"fmt"
	sakeyconfig "k8s-connectors/connectors/sakey/pkg/config"
	"k8s-connectors/pkg/config"
	"k8s-connectors/pkg/utils"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/yandex-cloud/go-sdk"

	connectorsv1 "k8s-connectors/connectors/sakey/api/v1"
	"k8s-connectors/connectors/sakey/controllers/phases"
)

// staticAccessKeyReconciler reconciles a StaticAccessKey object
type staticAccessKeyReconciler struct {
	client.Client
	log    logr.Logger
	scheme *runtime.Scheme
	sdk    *ycsdk.SDK

	// phases that are to be invoked on this object
	// IsUpdated blocks Update, and order of initializers matters,
	// thus if one of initializers fails, subsequent won't be processed.
	// Upon destruction of object, phase cleanups are called in
	// reverse order.
	phases []phases.AWSAccessKeyPhase
}

func NewAWSAccessKeyReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) (*staticAccessKeyReconciler, error) {
	sdk, err := ycsdk.Build(context.Background(), ycsdk.Config{
		Credentials: ycsdk.InstanceServiceAccount(),
	})
	if err != nil {
		return nil, err
	}
	return &staticAccessKeyReconciler{
		Client: client,
		log:    log,
		scheme: scheme,
		sdk:    sdk,
		phases: []phases.AWSAccessKeyPhase{
			// Register finalizer for the object (is blocked by allocation)
			&phases.FinalizerRegistrar{
				Client: &client,
			},
			// Allocate corresponding resource in cloud
			// (is blocked by finalizer registration,
			// because otherwise resource can leak)
			&phases.Allocator{
				Sdk:    sdk,
				Client: &client,
			},
			// In case spec was updated and our cloud resource does not match with
			// spec, we need to update cloud resource (is blocked by allocation)
			&phases.SpecMatcher{
				Sdk:    sdk,
				Client: &client,
			},
			// Update status of the object (is blocked by everything mutating)
			&phases.StatusUpdater{
				Sdk:    sdk,
				Client: &client,
			},
		},
	}, nil
}

//+kubebuilder:rbac:groups=connectors.cloud.yandex.ru,resources=staticaccesskeys,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=connectors.cloud.yandex.ru,resources=staticaccesskeys/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=connectors.cloud.yandex.ru,resources=staticaccesskeys/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=create;delete;list;get;watch;update
func (r *staticAccessKeyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithValues(sakeyconfig.LongName, req.NamespacedName)

	// Try to retrieve object from k8s
	var object connectorsv1.StaticAccessKey
	if err := r.Get(ctx, req.NamespacedName, &object); err != nil {
		// It still can be OK if we have not found it, and we do not need to reconcile it again

		// This outcome signifies that we just cannot find object, that is ok
		if apierrors.IsNotFound(err) {
			log.Info("object not found in k8s, reconciliation not possible")
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

	// Update all fragments of object, keeping track of whether
	// all of them are initialized
	for _, updater := range r.phases {
		isInitialized, err := updater.IsUpdated(ctx, &object)
		if err != nil {
			return config.GetErroredResult(err)
		}
		if !isInitialized {
			if err := updater.Update(ctx, log, &object); err != nil {
				return config.GetErroredResult(err)
			}
		}
	}

	return config.GetNormalResult()
}

func (r *staticAccessKeyReconciler) mustBeFinalized(object *connectorsv1.StaticAccessKey) (bool, error) {
	return !object.DeletionTimestamp.IsZero() && utils.ContainsString(object.Finalizers, sakeyconfig.FinalizerName), nil
}

func (r *staticAccessKeyReconciler) finalize(ctx context.Context, log logr.Logger, object *connectorsv1.StaticAccessKey) error {
	for i := len(r.phases); i != 0; i-- {
		if err := r.phases[i-1].Cleanup(ctx, log, object); err != nil {
			return fmt.Errorf("error during finalization: %v", err)
		}
	}
	log.Info("object finalized successfully")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *staticAccessKeyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&connectorsv1.StaticAccessKey{}).
		Complete(r)
}
