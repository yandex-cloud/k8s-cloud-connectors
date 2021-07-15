// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	sakey "k8s-connectors/connector/sakey/api/v1"
	sakeyconnector "k8s-connectors/connector/sakey/controller"
	sakeyconfig "k8s-connectors/connector/sakey/pkg/config"
	sakeywebhook "k8s-connectors/connector/sakey/webhook"
	ycr "k8s-connectors/connector/ycr/api/v1"
	ycrconnector "k8s-connectors/connector/ycr/controller"
	ycrconfig "k8s-connectors/connector/ycr/pkg/config"
	ycrwebhook "k8s-connectors/connector/ycr/webhook"
	ymq "k8s-connectors/connector/ymq/api/v1"
	ymqconnector "k8s-connectors/connector/ymq/controller"
	ymqconfig "k8s-connectors/connector/ymq/pkg/config"
	ymqwebhook "k8s-connectors/connector/ymq/webhook"
	yos "k8s-connectors/connector/yos/api/v1"
	yosconnector "k8s-connectors/connector/yos/controller"
	yosconfig "k8s-connectors/connector/yos/pkg/config"
	yoswebhook "k8s-connectors/connector/yos/webhook"
	"k8s-connectors/pkg/util"
	"k8s-connectors/pkg/webhook"

	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	// +kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()

	// Flag section
	metricsAddr          string
	enableLeaderElection bool
	debug                bool
	probeAddr            string
	clusterID            string
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(sakey.AddToScheme(scheme))
	utilruntime.Must(ycr.AddToScheme(scheme))
	utilruntime.Must(yos.AddToScheme(scheme))
	utilruntime.Must(ymq.AddToScheme(scheme))

	// Flag section
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&clusterID, "cluster-id", "", "ID of this cluster in the cloud.")
	flag.BoolVar(
		&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active connector manager.",
	)
	flag.BoolVar(&debug, "debug", false, "Enable debug logging for this connector manager.")
}

func getClusterIDFromNodeMetadata() (string, error) {
	ctx := context.Background()
	instanceIDReq, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"http://169.254.169.254:80/latest/meta-data/instance-id",
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("cannot make request to metadata: %w", err)
	}

	instanceIDResp, err := http.DefaultClient.Do(instanceIDReq)
	if err != nil {
		return "", fmt.Errorf("error while making request to metadata: %w", err)
	}

	if instanceIDResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unable to get instanceID from metadata")
	}

	instanceID, err := ioutil.ReadAll(instanceIDResp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read instanceID from metadata: %w", err)
	}

	// TODO (covariance) maybe we need to create one YCSDK and just pass it everywhere?
	yc, err := ycsdk.Build(
		ctx, ycsdk.Config{
			Credentials: ycsdk.InstanceServiceAccount(),
		},
	)
	if err != nil {
		return "", fmt.Errorf("unable to create ycsdk: %w", err)
	}

	instance, err := yc.Compute().Instance().Get(
		ctx, &compute.GetInstanceRequest{
			InstanceId: string(instanceID),
			View:       compute.InstanceView_BASIC,
		},
	)
	if err != nil {
		return "", fmt.Errorf("unable to get instance: %w", err)
	}

	ID, ok := instance.Labels["managed-kubernetes-cluster-id"]
	if !ok {
		return "", fmt.Errorf("instance does not have label with cluster id")
	}

	if err = instanceIDResp.Body.Close(); err != nil {
		return "", fmt.Errorf("unable to close response body: %w", err)
	}

	return ID, nil
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}

	log, err := util.NewZaprLogger(debug)
	if err != nil {
		fmt.Printf("unable to set up logger: %v", err)
		os.Exit(1)
	}
	ctrl.SetLogger(log)
	setupLog := ctrl.Log.WithName("setup")

	if err := execute(setupLog); err != nil {
		setupLog.Error(err, "connector manager error")
		os.Exit(1)
	}
}

func execute(log logr.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	if clusterID == "" {
		clusterID, err = getClusterIDFromNodeMetadata()
		if err != nil {
			return fmt.Errorf("unable to set cluster id: %w", err)
		}
	}

	mgr, err := ctrl.NewManager(
		ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:                 scheme,
			MetricsBindAddress:     metricsAddr,
			Port:                   9443,
			HealthProbeBindAddress: probeAddr,
			LeaderElection:         enableLeaderElection,
			LeaderElectionID:       "faeacf9e.cloud.yandex.com",
			CertDir:                "/etc/yandex-cloud-connectors/certs",
		},
	)
	if err != nil {
		return fmt.Errorf("unable to set up manager: %w", err)
	}

	if err := setupSAKeyConnector(ctx, log, mgr, clusterID); err != nil {
		return fmt.Errorf("unable to set up %s connector: %w", sakeyconfig.LongName, err)
	}
	if err := setupSAKeyWebhook(ctx, log, mgr); err != nil {
		return fmt.Errorf("unable to set up %s webhook: %w", sakeyconfig.LongName, err)
	}

	if err := setupYCRConnector(ctx, log, mgr, clusterID); err != nil {
		return fmt.Errorf("unable to set up %s connector: %w", ycrconfig.LongName, err)
	}
	if err := setupYCRWebhook(ctx, log, mgr); err != nil {
		return fmt.Errorf("unable to set up %s webhook: %w", ycrconfig.LongName, err)
	}

	if err := setupYMQConnector(ctx, log, mgr); err != nil {
		return fmt.Errorf("unable to set up %s connector: %w", ymqconfig.LongName, err)
	}
	if err := setupYMQWebhook(ctx, log, mgr); err != nil {
		return fmt.Errorf("unable to set up %s webhook: %w", ymqconfig.LongName, err)
	}

	if err := setupYOSConnector(ctx, log, mgr); err != nil {
		return fmt.Errorf("unable to set up %s connector: %w", yosconfig.LongName, err)
	}
	if err := setupYOSWebhook(ctx, log, mgr); err != nil {
		return fmt.Errorf("unable to set up %s webhook: %w", yosconfig.LongName, err)
	}

	// +kubebuilder:scaffold:builder

	log.V(1).Info("setting up health check")
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up health check: %w", err)
	}

	log.V(1).Info("setting up readiness check")
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up readiness check: %w", err)
	}

	log.V(1).Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("problem running manager: %w", err)
	}

	return nil
}

func setupSAKeyConnector(ctx context.Context, log logr.Logger, mgr ctrl.Manager, clusterID string) error {
	log.V(1).Info("starting " + sakeyconfig.ShortName + " connector")
	sakeyReconciler, err := sakeyconnector.NewStaticAccessKeyReconciler(
		ctx,
		mgr.GetClient(),
		ctrl.Log.WithName("connector").WithName(sakeyconfig.ShortName),
		clusterID,
	)
	if err != nil {
		return err
	}

	return sakeyReconciler.SetupWithManager(mgr)
}

func setupSAKeyWebhook(ctx context.Context, log logr.Logger, mgr ctrl.Manager) error {
	log.V(1).Info("starting " + sakeyconfig.ShortName + " webhook")

	validator, err := sakeywebhook.NewSAKeyValidator(ctx)
	if err != nil {
		return err
	}

	return webhook.RegisterValidatingHandler(mgr, &sakey.StaticAccessKey{}, validator)
}

func setupYCRConnector(ctx context.Context, log logr.Logger, mgr ctrl.Manager, clusterID string) error {
	log.V(1).Info("starting " + ycrconfig.ShortName + " connector")
	ycrReconciler, err := ycrconnector.NewYandexContainerRegistryReconciler(
		ctx,
		mgr.GetClient(),
		ctrl.Log.WithName("connector").WithName(ycrconfig.ShortName),
		clusterID,
	)
	if err != nil {
		return err
	}
	return ycrReconciler.SetupWithManager(mgr)
}

func setupYCRWebhook(ctx context.Context, log logr.Logger, mgr ctrl.Manager) error {
	log.V(1).Info("starting " + ycrconfig.ShortName + " webhook")

	validator, err := ycrwebhook.NewYCRValidator(ctx)
	if err != nil {
		return err
	}

	return webhook.RegisterValidatingHandler(mgr, &ycr.YandexContainerRegistry{}, validator)
}

func setupYMQConnector(_ context.Context, log logr.Logger, mgr ctrl.Manager) error {
	log.V(1).Info("starting " + ymqconfig.ShortName + " connector")
	ymqReconciler, err := ymqconnector.NewYandexMessageQueueReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("connector").WithName(ymqconfig.ShortName),
	)
	if err != nil {
		return err
	}
	return ymqReconciler.SetupWithManager(mgr)
}

func setupYMQWebhook(_ context.Context, log logr.Logger, mgr ctrl.Manager) error {
	log.V(1).Info("starting " + ymqconfig.ShortName + " webhook")

	validator, err := ymqwebhook.NewYMQValidator(mgr.GetClient())
	if err != nil {
		return err
	}

	return webhook.RegisterValidatingHandler(mgr, &ymq.YandexMessageQueue{}, validator)
}

func setupYOSConnector(_ context.Context, log logr.Logger, mgr ctrl.Manager) error {
	log.V(1).Info("starting " + yosconfig.ShortName + " connector")
	yosReconciler, err := yosconnector.NewYandexObjectStorageReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("connector").WithName(yosconfig.ShortName),
	)
	if err != nil {
		return err
	}
	return yosReconciler.SetupWithManager(mgr)
}

func setupYOSWebhook(_ context.Context, log logr.Logger, mgr ctrl.Manager) error {
	log.V(1).Info("starting " + yosconfig.ShortName + " webhook")

	validator, err := yoswebhook.NewYOSValidator(mgr.GetClient())
	if err != nil {
		return err
	}

	return webhook.RegisterValidatingHandler(mgr, &yos.YandexObjectStorage{}, validator)
}
