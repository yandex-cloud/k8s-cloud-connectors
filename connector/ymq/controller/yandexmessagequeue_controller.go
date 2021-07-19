// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "github.com/yandex-cloud/k8s-cloud-connectors/connector/ymq/api/v1"
	"github.com/yandex-cloud/k8s-cloud-connectors/connector/ymq/controller/adapter"
	ymqconfig "github.com/yandex-cloud/k8s-cloud-connectors/connector/ymq/pkg/config"
	ymqutils "github.com/yandex-cloud/k8s-cloud-connectors/connector/ymq/pkg/util"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/awsutils"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/config"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/phase"
)

// yandexMessageQueueReconciler reconciles a YandexContainerRegistry object
type yandexMessageQueueReconciler struct {
	client.Client
	adapter adapter.YandexMessageQueueAdapter
	log     logr.Logger
}

func NewYandexMessageQueueReconciler(cl client.Client, log logr.Logger) (*yandexMessageQueueReconciler, error) {
	impl, err := adapter.NewYandexMessageQueueAdapterSDK()
	if err != nil {
		return nil, err
	}
	return &yandexMessageQueueReconciler{
		Client:  cl,
		adapter: impl,
		log:     log,
	}, nil
}

// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=yandexmessagequeues,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=yandexmessagequeues/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=yandexmessagequeues/finalizers,verbs=update
// +kubebuilder:rbac:groups=connectors.cloud.yandex.com,resources=staticaccesskeys,verbs=get
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get

func (r *yandexMessageQueueReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithValues("name", req.NamespacedName)
	log.V(1).Info("started reconciliation")

	// Try to retrieve object from k8s
	var object connectorsv1.YandexMessageQueue
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
	sdk, err := ymqutils.NewSQSClient(ctx, cred)
	if err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to build sdk: %w", err))
	}

	// If object must be currently finalized, do it and quit
	if phase.MustBeFinalized(&object.ObjectMeta, ymqconfig.FinalizerName) {
		if err := r.finalize(ctx, log.WithName("finalize"), &object, sdk); err != nil {
			return config.GetErroredResult(fmt.Errorf("unable to finalize object: %w", err))
		}
		return config.GetNormalResult()
	}

	if err := phase.RegisterFinalizer(
		ctx, r.Client, log.WithName("register-finalizer"), &object.ObjectMeta, &object, ymqconfig.FinalizerName,
	); err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to register finalizer: %w", err))
	}

	if err := r.allocateResource(ctx, log.WithName("allocate-resource"), &object, sdk); err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to allocate resource: %w", err))
	}

	if err := r.matchSpec(ctx, log.WithName("match-spec"), &object, sdk); err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to match spec: %w", err))
	}

	if err := phase.ProvideConfigmap(
		ctx,
		r.Client,
		log.WithName("provide-configmap"),
		object.Name, ymqconfig.ShortName, object.Namespace,
		map[string]string{"URL": object.Status.QueueURL},
	); err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to provide configmap: %w", err))
	}

	log.V(1).Info("finished reconciliation")
	return config.GetNormalResult()
}

func (r *yandexMessageQueueReconciler) finalize(
	ctx context.Context,
	log logr.Logger,
	object *connectorsv1.YandexMessageQueue,
	sdk *sqs.SQS,
) error {
	log.V(1).Info("started")

	if err := phase.RemoveConfigmap(
		ctx,
		r.Client,
		log.WithName("remove-configmap"),
		object.Name, ymqconfig.ShortName, object.Namespace,
	); err != nil {
		return fmt.Errorf("unable to remove configmap: %w", err)
	}

	if err := r.deallocateResource(ctx, log.WithName("deallocate-resource"), object, sdk); err != nil {
		return fmt.Errorf("unable to deallocate resource: %w", err)
	}

	if err := phase.DeregisterFinalizer(
		ctx, r.Client, log.WithName("deregister-finalizer"), &object.ObjectMeta, object, ymqconfig.FinalizerName,
	); err != nil {
		return fmt.Errorf("unable to deregister finalizer: %w", err)
	}

	log.Info("successful")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *yandexMessageQueueReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&connectorsv1.YandexMessageQueue{}).
		Complete(r)
}
