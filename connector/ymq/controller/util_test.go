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

	sakey "k8s-connectors/connector/sakey/api/v1"
	connectorsv1 "k8s-connectors/connector/ymq/api/v1"
	"k8s-connectors/connector/ymq/controller/adapter"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
)

func setup(t *testing.T) (
	context.Context,
	logr.Logger,
	client.Client,
	adapter.YandexMessageQueueAdapter,
	yandexMessageQueueReconciler,
) {
	t.Helper()
	ad := adapter.NewFakeYandexMessageQueueAdapter()
	cl := k8sfake.NewFakeClient()
	log := logrfake.NewFakeLogger(t)
	return context.Background(), log, cl, ad, yandexMessageQueueReconciler{
		cl,
		ad,
		log,
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
	SAKey := sakey.StaticAccessKey{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: sakey.StaticAccessKeyStatus{
			SecretName: secretName,
		},
	}
	require.NoError(t, cl.Create(ctx, &SAKey))
}

func createDefaultQueue(name, namespace, key, queueName string) connectorsv1.YandexMessageQueue {
	return connectorsv1.YandexMessageQueue{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: connectorsv1.YandexMessageQueueSpec{
			Name:                          queueName,
			FifoQueue:                     false,
			ContentBasedDeduplication:     false,
			DelaySeconds:                  0,
			MaximumMessageSize:            262144,
			MessageRetentionPeriod:        60,
			ReceiveMessageWaitTimeSeconds: 0,
			VisibilityTimeout:             30,
			SAKeyName:                     key,
		},
	}
}
