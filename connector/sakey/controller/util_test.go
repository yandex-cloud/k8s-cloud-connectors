// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/api/v1"
	"github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/controller/adapter"
	k8sfake "github.com/yandex-cloud/k8s-cloud-connectors/testing/k8s-fake"
	logrfake "github.com/yandex-cloud/k8s-cloud-connectors/testing/logr-fake"
)

func setup(t *testing.T) (
	context.Context,
	logr.Logger,
	client.Client,
	adapter.StaticAccessKeyAdapter,
	staticAccessKeyReconciler,
) {
	t.Helper()
	ad := adapter.NewFakeStaticAccessKeyAdapter()
	cl := k8sfake.NewFakeClient()
	log := logrfake.NewFakeLogger(t)
	return context.Background(), log, cl, &ad, staticAccessKeyReconciler{
		cl,
		&ad,
		log,
		"test-cluster",
	}
}

func createObject(saID, metaName, namespace string) connectorsv1.StaticAccessKey {
	return connectorsv1.StaticAccessKey{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      metaName,
		},
		Spec: connectorsv1.StaticAccessKeySpec{
			ServiceAccountID: saID,
		},
	}
}
