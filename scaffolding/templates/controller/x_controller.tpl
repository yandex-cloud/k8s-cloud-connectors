package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "github.com/yandex-cloud/k8s-cloud-connectors/connector/{{ .shortName }}/api/v1"
	"github.com/yandex-cloud/k8s-cloud-connectors/connector/{{ .shortName }}/controller/adapter"
	{{ .shortName }}config "github.com/yandex-cloud/k8s-cloud-connectors/connector/{{ .shortName }}/pkg/config"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/config"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/phase"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/util"
)

// {{ .longName | untitle }}Reconciler reconciles a {{ .longName }} object
type {{ .longName | untitle }}Reconciler struct {
	client.Client
	adapter   adapter.{{ .longName }}Adapter
	log       logr.Logger
}

func New{{ .longName }}Reconciler(
	cl client.Client, log logr.Logger,
) (*{{ .longName | untitle }}Reconciler, error) {
	impl, err := adapter.New{{ .longName }}Adapter()
	if err != nil {
		return nil, err
	}
	return &{{ .longName | untitle }}Reconciler{
		Client:    cl,
		adapter:   impl,
		log:       log,
	}, nil
}

// +kubebuilder:rbac:groups={{ .groupName }},resources={{ .longName | lower }}s,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups={{ .groupName }},resources={{ .longName | lower }}s/status,verbs=get;update;patch
// +kubebuilder:rbac:groups={{ .groupName }},resources={{ .longName | lower }}s/finalizers,verbs=update

func (r *{{ .longName | untitle }}Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithValues("name", req.NamespacedName)
	log.V(1).Info("started reconciliation")

	// Try to retrieve object from k8s
	var object connectorsv1.{{ .longName }}
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
	mustBeFinalized, err := r.mustBeFinalized(&object)
	if err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to check if object must be finalized: %w", err))
	}
	if mustBeFinalized {
		if err := r.finalize(ctx, log.WithName("finalize"), &object); err != nil {
			return config.GetErroredResult(fmt.Errorf("unable to finalize object: %w", err))
		}
		return config.GetNormalResult()
	}

	if err := phase.RegisterFinalizer(
		ctx, r.Client, log.WithName("register-finalizer"), &object.ObjectMeta, &object, {{ .shortName }}config.FinalizerName,
	); err != nil {
		return config.GetErroredResult(fmt.Errorf("unable to register finalizer: %w", err))
	}

	log.V(1).Info("finished reconciliation")
	return config.GetNormalResult()
}

func (r *{{ .longName | untitle }}Reconciler) mustBeFinalized(object *connectorsv1.{{ .longName }}) (bool, error) {
	return !object.DeletionTimestamp.IsZero() && util.ContainsString(object.Finalizers, {{ .shortName }}config.FinalizerName), nil
}

func (r *{{ .longName | untitle }}Reconciler) finalize(
	ctx context.Context, log logr.Logger, object *connectorsv1.{{ .longName }},
) error {
	log.V(1).Info("started")

	if err := phase.DeregisterFinalizer(
		ctx, r.Client, log.WithName("deregister-finalizer"), &object.ObjectMeta, object, {{ .shortName }}config.FinalizerName,
	); err != nil {
		return fmt.Errorf("unable to deregister finalizer: %w", err)
	}

	log.Info("successful")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *{{ .longName | untitle }}Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&connectorsv1.{{ .longName }}{}).
		Complete(r)
}
