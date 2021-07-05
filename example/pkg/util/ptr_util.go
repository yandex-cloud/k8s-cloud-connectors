// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package util

func StrPtr(v string) *string {
	return &v
}

func Int64Ptr(v int64) *int64 {
	return &v
}
