// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type ObjectInjector interface {
	InjectObject(obj runtime.Object) error
}
