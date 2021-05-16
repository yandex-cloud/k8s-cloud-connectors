// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package logr_fake

import (
	"github.com/go-logr/logr"
	"testing"
)

type FakeLogger struct {
	Level int
	Name  string
	KV    []interface{}
	T     *testing.T
}

func NewFakeLogger(t *testing.T) logr.Logger {
	return FakeLogger{
		Level: 0,
		Name:  "",
		KV:    []interface{}{},
		T:     t,
	}
}

// Enabled tests whether this Logger is enabled.  For example, commandline
// flags might be used to set the logging verbosity and disable some info
// logs.
func (r FakeLogger) Enabled() bool {
	return true
}

// Info logs a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func (r FakeLogger) Info(msg string, keysAndValues ...interface{}) {
	r.T.Log("level", r.Level, "info:")
	r.T.Log(r.KV...)
	r.T.Log(msg)
	r.T.Log(keysAndValues...)
}

// Error logs an error, with the given message and key/value pairs as context.
// It functions similarly to calling Info with the "error" named value, but may
// have unique behavior, and should be preferred for logging errors (see the
// package documentations for more information).
//
// The msg field should be used to add context to any underlying error,
// while the err field should be used to attach the actual error that
// triggered this log line, if present.
func (r FakeLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	r.T.Log("level", r.Level, "error:")
	r.T.Log(r.KV...)
	r.T.Log(err)
	r.T.Log(msg)
	r.T.Log(keysAndValues...)
}

// V returns an Logger value for a specific verbosity level, relative to
// this Logger.  In other words, V values are additive.  V higher verbosity
// level means a log message is less important.  It's illegal to pass a log
// level less than zero.
func (r FakeLogger) V(level int) logr.Logger {
	return FakeLogger{
		Level: r.Level + level,
		Name:  r.Name,
		KV:    r.KV,
		T:     r.T,
	}
}

// WithValues adds some key-value pairs of context to a logger.
// See Info for documentation on how key/value pairs work.
func (r FakeLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return FakeLogger{
		Level: r.Level,
		Name:  r.Name,
		KV:    append(r.KV, keysAndValues...),
		T:     r.T,
	}
}

// WithName adds a new element to the logger's name.
// Successive calls with WithName continue to append
// suffixes to the logger's name.  It's strongly recommended
// that name segments contain only letters, digits, and hyphens
// (see the package documentation for more information).
func (r FakeLogger) WithName(name string) logr.Logger {
	return FakeLogger{
		Level: r.Level,
		Name:  r.Name + name,
		KV:    r.KV,
		T:     r.T,
	}
}
