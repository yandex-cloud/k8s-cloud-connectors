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
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	sakey "k8s-connectors/connector/sakey/api/v1"
	sakeyconnector "k8s-connectors/connector/sakey/controller"
	sakeyconfig "k8s-connectors/connector/sakey/pkg/config"
	ycr "k8s-connectors/connector/ycr/api/v1"
	ycrconnector "k8s-connectors/connector/ycr/controller"
	ycrconfig "k8s-connectors/connector/ycr/pkg/config"
	ymq "k8s-connectors/connector/ymq/api/v1"
	ymqconnector "k8s-connectors/connector/ymq/controller"
	ymqconfig "k8s-connectors/connector/ymq/pkg/config"
	yos "k8s-connectors/connector/yos/api/v1"
	yosconnector "k8s-connectors/connector/yos/controller"
	yosconfig "k8s-connectors/connector/yos/pkg/config"
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
		return "", fmt.Errorf("cannot make request to metadata: %v", err)
	}

	instanceIDResp, err := http.DefaultClient.Do(instanceIDReq)
	if err != nil {
		return "", fmt.Errorf("error while making request to metadata: %v", err)
	}

	if instanceIDResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unable to get instanceID from metadata")
	}

	instanceID, err := ioutil.ReadAll(instanceIDResp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read instanceID from metadata: %v", err)
	}

	// TODO (covariance) maybe we need to create one YCSDK and just pass it everywhere?
	yc, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: ycsdk.InstanceServiceAccount(),
	})
	if err != nil {
		return "", fmt.Errorf("unable to create ycsdk: %v", err)
	}

	instance, err := yc.Compute().Instance().Get(ctx, &compute.GetInstanceRequest{
		InstanceId: string(instanceID),
		View:       compute.InstanceView_BASIC,
	})
	if err != nil {
		return "", fmt.Errorf("unable to get instance: %v", err)
	}

	clusterID, ok := instance.Labels["managed-kubernetes-cluster-id"]
	if !ok {
		return "", fmt.Errorf("instance does not have label with cluster id")
	}

	if err = instanceIDResp.Body.Close(); err != nil {
		return "", fmt.Errorf("unable to close response body: %v", err)
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
	var probeAddr string
	var clusterID string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&clusterID, "cluster-id", "", "ID of this cluster in the cloud")
	flag.BoolVar(
		&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.",
	)
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("starting manager setup")

	if clusterID == "" {
		var err error
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

	setupLog.Info("setting up health check")
	err = mgr.AddHealthzCheck("healthz", healthz.Ping)
	setupErrorExit(err, "health check", setupLog)

	setupLog.Info("setting up readiness check")
	err = mgr.AddReadyzCheck("readyz", healthz.Ping)
	setupErrorExit(err, "readiness check", setupLog)

	setupLog.Info("starting manager")
	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func setupSAKeyConnector(mgr ctrl.Manager, clusterID string) {
	setupLog.Info("starting " + sakeyconfig.ShortName + " connector")
	sakeyReconciler, err := sakeyconnector.NewStaticAccessKeyReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName(sakeyconfig.LongName),
		clusterID,
	)
	controllerCreationErrorExit(err, sakeyconfig.LongName, setupLog)
	err = sakeyReconciler.SetupWithManager(mgr)
	controllerCreationErrorExit(err, sakeyconfig.LongName, setupLog)
}

func setupSAKeyWebhook(mgr ctrl.Manager) {
	setupLog.Info("starting " + sakeyconfig.ShortName + " webhook")
	err := (&sakey.StaticAccessKey{}).SetupWebhookWithManager(mgr)
	webhookCreationErrorExit(err, sakeyconfig.LongName, setupLog)
}

func setupYCRConnector(mgr ctrl.Manager, clusterID string) {
	setupLog.Info("starting " + ycrconfig.ShortName + " connector")
	ycrReconciler, err := ycrconnector.NewYandexContainerRegistryReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName(ycrconfig.LongName),
		clusterID,
	)
	controllerCreationErrorExit(err, ycrconfig.LongName, setupLog)
	err = ycrReconciler.SetupWithManager(mgr)
	controllerCreationErrorExit(err, ycrconfig.LongName, setupLog)
}

func setupYCRWebhook(mgr ctrl.Manager) {
	setupLog.Info("starting " + ycrconfig.ShortName + " webhook")
	err := (&ycr.YandexContainerRegistry{}).SetupWebhookWithManager(mgr)
	webhookCreationErrorExit(err, ycrconfig.LongName, setupLog)
}

func setupYMQConnector(mgr ctrl.Manager) {
	setupLog.Info("starting " + ymqconfig.ShortName + " connector")
	ymqReconciler, err := ymqconnector.NewYandexMessageQueueReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName(ymqconfig.LongName),
	)
	controllerCreationErrorExit(err, ymqconfig.LongName, setupLog)
	err = ymqReconciler.SetupWithManager(mgr)
	controllerCreationErrorExit(err, ymqconfig.LongName, setupLog)
}

func setupYMQWebhook(mgr ctrl.Manager) {
	setupLog.Info("starting " + ymqconfig.ShortName + " webhook")
	err := (&ymq.YandexMessageQueue{}).SetupWebhookWithManager(mgr)
	webhookCreationErrorExit(err, ymqconfig.LongName, setupLog)
}

func setupYOSConnector(mgr ctrl.Manager) {
	setupLog.Info("starting " + yosconfig.ShortName + " connector")
	yosReconciler, err := yosconnector.NewYandexObjectStorageReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName(yosconfig.LongName),
	)
	controllerCreationErrorExit(err, yosconfig.LongName, setupLog)
	err = yosReconciler.SetupWithManager(mgr)
	controllerCreationErrorExit(err, yosconfig.LongName, setupLog)
}

func setupYOSWebhook(mgr ctrl.Manager) {
	setupLog.Info("starting " + yosconfig.ShortName + " webhook")
	err := (&yos.YandexObjectStorage{}).SetupWebhookWithManager(mgr)
	webhookCreationErrorExit(err, yosconfig.LongName, setupLog)
}
