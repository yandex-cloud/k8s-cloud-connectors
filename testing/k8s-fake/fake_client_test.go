// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package k8s_fake

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestCreation(t *testing.T) {
	var _ client.Client
	_ = FakeClient{}
}
