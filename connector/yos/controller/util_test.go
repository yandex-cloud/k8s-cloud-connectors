// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "k8s-connectors/connector/sakey/api/v1"
	v12 "k8s-connectors/connector/yos/api/v1"
	"k8s-connectors/connector/yos/controller/adapter"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
)

func setup(t *testing.T) (
	context.Context,
	logr.Logger,
	client.Client,
	adapter.YandexObjectStorageAdapter,
	yandexObjectStorageReconciler,
) {
	t.Helper()
	ad := adapter.NewFakeYandexObjectStorageAdapter()
	cl := k8sfake.NewFakeClient()
	log := logrfake.NewFakeLogger(t)
	return context.Background(), log, cl, ad, yandexObjectStorageReconciler{
		cl,
		ad,
		log,
	}
}

func createObject(name, sakey, acl, metaName, namespace string) v12.YandexObjectStorage {
	return v12.YandexObjectStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metaName,
			Namespace: namespace,
		},
		Spec: v12.YandexObjectStorageSpec{
			Name:      name,
			ACL:       acl,
			SAKeyName: sakey,
		},
	}
}

func createSAKeyRequireNoError(ctx context.Context, t *testing.T, cl client.Client, name, namespace string) {
	t.Helper()
	secretName := name + "/" + namespace + "/secret"
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"key":    []byte("test-key"),
			"secret": []byte("test-secret"),
		},
	}
	require.NoError(t, cl.Create(ctx, &secret))
	sakey := connectorsv1.StaticAccessKey{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: connectorsv1.StaticAccessKeyStatus{
			SecretName: secretName,
		},
	}
	require.NoError(t, cl.Create(ctx, &sakey))
}
