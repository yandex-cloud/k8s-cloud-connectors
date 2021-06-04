// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package util

import (
	"strconv"
)

func IntToStringPtr(v int) *string {
	s := strconv.Itoa(v)
	return &s
}

func StringPtr(s string) *string { return &s }
