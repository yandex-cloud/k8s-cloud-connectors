// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package errorhandling

import (
	"fmt"
)

type ConnectorError struct {
	message  string
	code     string
	original error
}

func New(message, code string, initial error) ConnectorError {
	return ConnectorError{
		message:  message,
		code:     code,
		original: initial,
	}
}

func (r ConnectorError) Error() string {
	return fmt.Sprintf("%s: %v", r.message, r.original)
}

func (r ConnectorError) Code() string {
	return r.code
}

func (r ConnectorError) Initial() error {
	return r.original
}
