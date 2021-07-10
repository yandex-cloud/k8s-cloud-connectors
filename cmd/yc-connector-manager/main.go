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

	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

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
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(sakey.AddToScheme(scheme))
	utilruntime.Must(ycr.AddToScheme(scheme))
	utilruntime.Must(yos.AddToScheme(scheme))
	utilruntime.Must(ymq.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func getClusterIDFromNodeMetadata() (string, error) {
	ctx := context.Background()
	instanceIDReq, err := http.NewRequestWithContext(ctx,
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
	yc, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: ycsdk.InstanceServiceAccount(),
	})
	if err != nil {
		return "", fmt.Errorf("unable to create ycsdk: %w", err)
	}

	instance, err := yc.Compute().Instance().Get(ctx, &compute.GetInstanceRequest{
		InstanceId: string(instanceID),
		View:       compute.InstanceView_BASIC,
	})
	if err != nil {
		return "", fmt.Errorf("unable to get instance: %w", err)
	}

	clusterID, ok := instance.Labels["managed-kubernetes-cluster-id"]
	if !ok {
		return "", fmt.Errorf("instance does not have label with cluster id")
	}

	if err = instanceIDResp.Body.Close(); err != nil {
		return "", fmt.Errorf("unable to close response body: %w", err)
	}

	return clusterID, nil
}

func setupErrorExit(err error, setupEntity string, log logr.Logger) {
	if err != nil {
		log.Error(err, "unable to setup "+setupEntity)
		os.Exit(1)
	}
}

func controllerCreationErrorExit(err error, controllerName string, log logr.Logger) {
	setupErrorExit(err, "controller "+controllerName, log)
}

func webhookCreationErrorExit(err error, webhookName string, log logr.Logger) {
	setupErrorExit(err, "webhook "+webhookName, log)
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var debug bool
	var probeAddr string
	var clusterID string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&clusterID, "cluster-id", "", "ID of this cluster in the cloud.")
	flag.BoolVar(
		&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active connector manager.",
	)
	flag.BoolVar(&debug, "debug", false, "Enable debug logging for this connector manager.")
	flag.Parse()

	log, err := util.NewZaprLogger(debug)
	if err != nil {
		fmt.Printf("unable to set up logger: %v", err)
		os.Exit(1)
	}
	ctrl.SetLogger(log)
	setupLog.Info("starting manager setup")

	if clusterID == "" {
		clusterID, err = getClusterIDFromNodeMetadata()
		if err != nil {
			setupLog.Error(err, "unable to set cluster id")
			os.Exit(1)
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
	setupErrorExit(err, "manager", setupLog)

	setupSAKeyConnector(mgr, clusterID)
	setupSAKeyWebhook(mgr)

	setupYCRConnector(mgr, clusterID)
	setupYCRWebhook(mgr)

	setupYMQConnector(mgr)
	setupYMQWebhook(mgr)

	setupYOSConnector(mgr)
	setupYOSWebhook(mgr)

	// +kubebuilder:scaffold:builder

	setupLog.V(1).Info("setting up health check")
	err = mgr.AddHealthzCheck("healthz", healthz.Ping)
	setupErrorExit(err, "health check", setupLog)

	setupLog.V(1).Info("setting up readiness check")
	err = mgr.AddReadyzCheck("readyz", healthz.Ping)
	setupErrorExit(err, "readiness check", setupLog)

	setupLog.V(1).Info("starting manager")
	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func setupSAKeyConnector(mgr ctrl.Manager, clusterID string) {
	setupLog.V(1).Info("starting " + sakeyconfig.ShortName + " connector")
	sakeyReconciler, err := sakeyconnector.NewStaticAccessKeyReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("connector").WithName(sakeyconfig.ShortName),
		clusterID,
	)
	controllerCreationErrorExit(err, sakeyconfig.LongName, setupLog)
	controllerCreationErrorExit(sakeyReconciler.SetupWithManager(mgr), sakeyconfig.LongName, setupLog)
}

func setupSAKeyWebhook(mgr ctrl.Manager) {
	setupLog.V(1).Info("starting " + sakeyconfig.ShortName + " webhook")

	webhookCreationErrorExit(
		webhook.RegisterValidatingHandler(mgr, &sakey.StaticAccessKey{}, &sakeywebhook.SAKeyValidator{}),
		sakeyconfig.LongName,
		setupLog,
	)
}

func setupYCRConnector(mgr ctrl.Manager, clusterID string) {
	setupLog.V(1).Info("starting " + ycrconfig.ShortName + " connector")
	ycrReconciler, err := ycrconnector.NewYandexContainerRegistryReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("connector").WithName(ycrconfig.ShortName),
		clusterID,
	)
	controllerCreationErrorExit(err, ycrconfig.LongName, setupLog)
	controllerCreationErrorExit(ycrReconciler.SetupWithManager(mgr), ycrconfig.LongName, setupLog)
}

func setupYCRWebhook(mgr ctrl.Manager) {
	setupLog.V(1).Info("starting " + ycrconfig.ShortName + " webhook")

	webhookCreationErrorExit(
		webhook.RegisterValidatingHandler(mgr, &ycr.YandexContainerRegistry{}, &ycrwebhook.YCRValidator{}),
		ycrconfig.LongName,
		setupLog,
	)
}

func setupYMQConnector(mgr ctrl.Manager) {
	setupLog.V(1).Info("starting " + ymqconfig.ShortName + " connector")
	ymqReconciler, err := ymqconnector.NewYandexMessageQueueReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("connector").WithName(ymqconfig.ShortName),
	)
	controllerCreationErrorExit(err, ymqconfig.LongName, setupLog)
	controllerCreationErrorExit(ymqReconciler.SetupWithManager(mgr), ymqconfig.LongName, setupLog)
}

func setupYMQWebhook(mgr ctrl.Manager) {
	setupLog.V(1).Info("starting " + ymqconfig.ShortName + " webhook")

	webhookCreationErrorExit(
		webhook.RegisterValidatingHandler(mgr, &ymq.YandexMessageQueue{}, &ymqwebhook.YMQValidator{}),
		ymqconfig.LongName,
		setupLog,
	)
}

func setupYOSConnector(mgr ctrl.Manager) {
	setupLog.V(1).Info("starting " + yosconfig.ShortName + " connector")
	yosReconciler, err := yosconnector.NewYandexObjectStorageReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("connector").WithName(yosconfig.ShortName),
	)
	controllerCreationErrorExit(err, yosconfig.LongName, setupLog)
	controllerCreationErrorExit(yosReconciler.SetupWithManager(mgr), yosconfig.LongName, setupLog)
}

func setupYOSWebhook(mgr ctrl.Manager) {
	setupLog.V(1).Info("starting " + yosconfig.ShortName + " webhook")

	webhookCreationErrorExit(
		webhook.RegisterValidatingHandler(mgr, &yos.YandexObjectStorage{}, &yoswebhook.YOSValidator{}),
		yosconfig.LongName,
		setupLog,
	)
}
