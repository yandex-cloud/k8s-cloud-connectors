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

	sakey "github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/api/v1"
	sakeyconnector "github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/controller"
	sakeyconfig "github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/pkg/config"
	sakeywebhook "github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/webhook"
	ycr "github.com/yandex-cloud/k8s-cloud-connectors/connector/ycr/api/v1"
	ycrconnector "github.com/yandex-cloud/k8s-cloud-connectors/connector/ycr/controller"
	ycrconfig "github.com/yandex-cloud/k8s-cloud-connectors/connector/ycr/pkg/config"
	ycrwebhook "github.com/yandex-cloud/k8s-cloud-connectors/connector/ycr/webhook"
	ymq "github.com/yandex-cloud/k8s-cloud-connectors/connector/ymq/api/v1"
	ymqconnector "github.com/yandex-cloud/k8s-cloud-connectors/connector/ymq/controller"
	ymqconfig "github.com/yandex-cloud/k8s-cloud-connectors/connector/ymq/pkg/config"
	ymqwebhook "github.com/yandex-cloud/k8s-cloud-connectors/connector/ymq/webhook"
	yos "github.com/yandex-cloud/k8s-cloud-connectors/connector/yos/api/v1"
	yosconnector "github.com/yandex-cloud/k8s-cloud-connectors/connector/yos/controller"
	yosconfig "github.com/yandex-cloud/k8s-cloud-connectors/connector/yos/pkg/config"
	yoswebhook "github.com/yandex-cloud/k8s-cloud-connectors/connector/yos/webhook"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/util"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/webhook"

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

func getClusterIDFromNodeMetadata(sdk *ycsdk.SDK) (string, error) {
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

	instance, err := sdk.Compute().Instance().Get(
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

//nolint:gocyclo
func execute(log logr.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sdk, err := ycsdk.Build(
		ctx, ycsdk.Config{
			Credentials: ycsdk.InstanceServiceAccount(),
		},
	)
	if err != nil {
		return fmt.Errorf("unable to create ycsdk: %w", err)
	}

	if clusterID == "" {
		clusterID, err = getClusterIDFromNodeMetadata(sdk)
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

	if err := setupSAKeyConnector(log, mgr, sdk, clusterID); err != nil {
		return fmt.Errorf("unable to set up %s connector: %w", sakeyconfig.LongName, err)
	}
	if err := setupSAKeyWebhook(log, mgr, sdk); err != nil {
		return fmt.Errorf("unable to set up %s webhook: %w", sakeyconfig.LongName, err)
	}

	if err := setupYCRConnector(log, mgr, sdk, clusterID); err != nil {
		return fmt.Errorf("unable to set up %s connector: %w", ycrconfig.LongName, err)
	}
	if err := setupYCRWebhook(log, mgr, sdk); err != nil {
		return fmt.Errorf("unable to set up %s webhook: %w", ycrconfig.LongName, err)
	}

	if err := setupYMQConnector(log, mgr); err != nil {
		return fmt.Errorf("unable to set up %s connector: %w", ymqconfig.LongName, err)
	}
	if err := setupYMQWebhook(log, mgr); err != nil {
		return fmt.Errorf("unable to set up %s webhook: %w", ymqconfig.LongName, err)
	}

	if err := setupYOSConnector(log, mgr); err != nil {
		return fmt.Errorf("unable to set up %s connector: %w", yosconfig.LongName, err)
	}
	if err := setupYOSWebhook(log, mgr); err != nil {
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

func setupSAKeyConnector(log logr.Logger, mgr ctrl.Manager, sdk *ycsdk.SDK, clusterID string) error {
	log.V(1).Info("starting " + sakeyconfig.ShortName + " connector")
	sakeyReconciler := sakeyconnector.NewStaticAccessKeyReconciler(
		ctrl.Log.WithName("connector").WithName(sakeyconfig.ShortName),
		mgr.GetClient(),
		sdk,
		clusterID,
	)
	return sakeyReconciler.SetupWithManager(mgr)
}

func setupSAKeyWebhook(log logr.Logger, mgr ctrl.Manager, sdk *ycsdk.SDK) error {
	log.V(1).Info("starting " + sakeyconfig.ShortName + " webhook")
	validator := sakeywebhook.NewSAKeyValidator(sdk)
	return webhook.RegisterValidatingHandler(mgr, &sakey.StaticAccessKey{}, validator)
}

func setupYCRConnector(log logr.Logger, mgr ctrl.Manager, sdk *ycsdk.SDK, clusterID string) error {
	log.V(1).Info("starting " + ycrconfig.ShortName + " connector")
	ycrReconciler := ycrconnector.NewYandexContainerRegistryReconciler(
		ctrl.Log.WithName("connector").WithName(ycrconfig.ShortName),
		mgr.GetClient(),
		sdk,
		clusterID,
	)
	return ycrReconciler.SetupWithManager(mgr)
}

func setupYCRWebhook(log logr.Logger, mgr ctrl.Manager, sdk *ycsdk.SDK) error {
	log.V(1).Info("starting " + ycrconfig.ShortName + " webhook")
	validator := ycrwebhook.NewYCRValidator(sdk)
	return webhook.RegisterValidatingHandler(mgr, &ycr.YandexContainerRegistry{}, validator)
}

func setupYMQConnector(log logr.Logger, mgr ctrl.Manager) error {
	log.V(1).Info("starting " + ymqconfig.ShortName + " connector")
	ymqReconciler := ymqconnector.NewYandexMessageQueueReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("connector").WithName(ymqconfig.ShortName),
	)
	return ymqReconciler.SetupWithManager(mgr)
}

func setupYMQWebhook(log logr.Logger, mgr ctrl.Manager) error {
	log.V(1).Info("starting " + ymqconfig.ShortName + " webhook")
	validator := ymqwebhook.NewYMQValidator(mgr.GetClient())
	return webhook.RegisterValidatingHandler(mgr, &ymq.YandexMessageQueue{}, validator)
}

func setupYOSConnector(log logr.Logger, mgr ctrl.Manager) error {
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

func setupYOSWebhook(log logr.Logger, mgr ctrl.Manager) error {
	log.V(1).Info("starting " + yosconfig.ShortName + " webhook")

	validator, err := yoswebhook.NewYOSValidator(mgr.GetClient())
	if err != nil {
		return err
	}

	return webhook.RegisterValidatingHandler(mgr, &yos.YandexObjectStorage{}, validator)
}
