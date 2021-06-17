// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package util

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

const developmentStepsSkipped = 1

func NewZaprLogger(debug bool) (logr.Logger, error) {
	var log *zap.Logger
	var err error
	if debug {
		log, err = zap.NewDevelopment(zap.AddCallerSkip(developmentStepsSkipped))
	} else {
		log, err = zap.NewProduction()
	}
	if err != nil {
		return nil, err
	}
	return zapr.NewLogger(log), nil
}
