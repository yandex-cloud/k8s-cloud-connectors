// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package util

func EqualsStringString(lhs, rhs map[string]string) bool {
	for k, v := range lhs {
		if other, ok := rhs[k]; !ok || v != other {
			return false
		}
	}

	for k, v := range rhs {
		if other, ok := lhs[k]; !ok || v != other {
			return false
		}
	}

	return true
}
