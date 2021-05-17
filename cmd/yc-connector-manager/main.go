// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	sakey "k8s-connectors/connectors/sakey/api/v1"
	ycr "k8s-connectors/connectors/ycr/api/v1"
	ymq "k8s-connectors/connectors/ymq/api/v1"
	yos "k8s-connectors/connectors/yos/api/v1"

	sakeyconfig "k8s-connectors/connectors/sakey/pkg/config"
	ycrconfig "k8s-connectors/connectors/ycr/pkg/config"
	ymqconfig "k8s-connectors/connectors/ymq/pkg/config"
	yosconfig "k8s-connectors/connectors/yos/pkg/config"

	sakeyconnector "k8s-connectors/connectors/sakey/controllers"
	ycrconnector "k8s-connectors/connectors/ycr/controllers"
	ymqconnector "k8s-connectors/connectors/ymq/controllers"
	yosconnector "k8s-connectors/connectors/yos/controllers"
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

func setupErrorExit(err error, setupEntity string) {
	if err != nil {
		setupLog.Error(err, "unable to setup "+setupEntity)
		os.Exit(1)
	}
}

func controllerCreationErrorExit(err error, controllerName string) {
	setupErrorExit(err, "controller "+controllerName)
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
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
	setupErrorExit(err, "manager")

	sakeyReconciler, err := sakeyconnector.NewStaticAccessKeyReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName(sakeyconfig.LongName),
		mgr.GetScheme(),
	)
	controllerCreationErrorExit(err, sakeyconfig.LongName)
	err = sakeyReconciler.SetupWithManager(mgr)
	controllerCreationErrorExit(err, sakeyconfig.LongName)

	ycrReconciler, err := ycrconnector.NewYandexContainerRegistryReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName(ycrconfig.LongName),
		mgr.GetScheme(),
	)
	controllerCreationErrorExit(err, ycrconfig.LongName)
	err = ycrReconciler.SetupWithManager(mgr)
	controllerCreationErrorExit(err, ycrconfig.LongName)

	ymqReconciler, err := ymqconnector.NewYandexMessageQueueReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName(ymqconfig.LongName),
		mgr.GetScheme(),
	)
	controllerCreationErrorExit(err, ymqconfig.LongName)
	err = ymqReconciler.SetupWithManager(mgr)
	controllerCreationErrorExit(err, ymqconfig.LongName)

	yosReconciler, err := yosconnector.NewYandexObjectStorageReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName(yosconfig.LongName),
		mgr.GetScheme(),
	)
	controllerCreationErrorExit(err, yosconfig.LongName)
	err = yosReconciler.SetupWithManager(mgr)
	controllerCreationErrorExit(err, yosconfig.LongName)

	// +kubebuilder:scaffold:builder

	err = mgr.AddHealthzCheck("healthz", healthz.Ping)
	setupErrorExit(err, "health check")

	err = mgr.AddReadyzCheck("readyz", healthz.Ping)
	setupErrorExit(err, "readiness check")

	setupLog.Info("starting manager")
	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
