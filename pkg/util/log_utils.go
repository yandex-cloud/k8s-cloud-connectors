// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package util

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const stepsSkipped = 1

func NewZaprLogger(debug bool) (logr.Logger, error) {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	if debug {
		loggerConfig.Development = true
		loggerConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	log, err := loggerConfig.Build(zap.AddCallerSkip(stepsSkipped))
	if err != nil {
		return nil, err
	}
	return zapr.NewLogger(log), nil
}
