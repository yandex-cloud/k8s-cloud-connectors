// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phase

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
)

func setup(t *testing.T) (context.Context, logr.Logger, client.Client) {
	t.Helper()
	return context.Background(), logrfake.NewFakeLogger(t), k8sfake.NewFakeClient()
}
