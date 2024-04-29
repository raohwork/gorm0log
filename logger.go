// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package gorm0log provides a gorm logger using zerolog as backend.
//
// Thanks to zerolog, we have goo performance since most features are skipped if the
// log message is not visible.
//
// If you queries gorm with context ([gorm.DB.WithContext]), the context can be
// handled by a function in [Config] named "Customize".
//
// Default value of [Config] should be fairly enough for small projects.
//
// [gorm.DB.Debug] switches underlying [zerolog.Logger] to Debug level, which shows
// every message but Trace level ones. Some features, controlled by [Config], can be
// set to customized level. So you can disable those messages by changing its level
// to [UseTrace].
package gorm0log

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm/logger"
)

// DefaultLogLevelMap enables sql dumping if you use [logger.Info] log mode.
//
// It translates Gorm's [logger.Info] mode to [zerolog.DebugLevel], so sql dumping
// is logged.
func DefaultLogLevelMap(i logger.LogLevel) zerolog.Level {
	switch {
	case i <= logger.Silent:
		return zerolog.Disabled
	case i == logger.Error:
		return zerolog.ErrorLevel
	case i == logger.Warn:
		return zerolog.WarnLevel
	}
	return zerolog.DebugLevel
}

// WithTimeTracking enables time tracking info if you use [logger.Info] log mode.
//
// It translates Gorm's [logger.Info] mode to [zerolog.TraceLevel], so sql dumping
// and time tracking info are logged.
func WithTimeTracking(i logger.LogLevel) zerolog.Level {
	switch {
	case i <= logger.Silent:
		return zerolog.Disabled
	case i == logger.Error:
		return zerolog.ErrorLevel
	case i == logger.Warn:
		return zerolog.WarnLevel
	}
	return zerolog.TraceLevel
}

// Logger implements [logger.Interface].
//
// This implementation provides some customizable features, take a look at [Config].
//
// Using [gorm.DB.Debug] switches log level of zerolog.Logger to Debug level.
//
// Zero value means a logger that:
//
//   - Log to empty [zerolog.Logger], which logs nothing.
//   - Using default [Config].
type Logger struct {
	zerolog.Logger
	Config
}

// LogMode implements [logger.Interface], to control which message is visible.
func (l *Logger) LogMode(lv logger.LogLevel) logger.Interface {
	var lvl zerolog.Level
	switch {
	case lv <= logger.Silent:
		lvl = zerolog.Disabled
	case lv == logger.Error:
		lvl = zerolog.ErrorLevel
	case lv == logger.Warn:
		lvl = zerolog.WarnLevel
	default:
		lvl = zerolog.DebugLevel
	}

	return &Logger{
		Logger: l.Logger.Level(lvl),
		Config: l.Config,
	}
}

// Info implements [logger.Interface], to show a message at Info level.
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.Logger.Info().Func(l.custom(ctx)).Msgf(msg, args...)
}

// Warn implements [logger.Interface], to show a message at Warn level.
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.Logger.Warn().Func(l.custom(ctx)).Msgf(msg, args...)
}

// Error implements [logger.Interface], to show a message at Error level.
func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.Logger.Error().Func(l.custom(ctx)).Msgf(msg, args...)
}

// Trace implements [logger.Ingerface]. It is called every query by Gorm, so we can
// provide useful features like slow log or sql dump.
func (l *Logger) Trace(ctx context.Context, begin time.Time, f func() (string, int64), err error) {
	dur := time.Since(begin)

	if err != nil {
		ev := l.errLevel(err, l.Logger)
		ev.Func(l.custom(ctx)).Func(l.logErr(err, f)).Msg("a sql error occurred")

		if ev.Enabled() {
			// do not log other messages
			return
		}
	}

	if l.SlowThreshold > 0 && dur >= l.SlowThreshold {
		// slow log
		l.slowLevel(l.Logger).
			Func(l.custom(ctx)).
			Func(l.logSlow(dur, f)).
			Msg("sql query time exceeds threshold")
		return
	}

	l.dumpLevel(l.Logger).
		Func(l.custom(ctx)).
		Func(l.logDump(dur, f)).
		Msg("dump sql")
}

// ParamsFilter implements [gorm.ParamsFilter] to check if parameters should be shown.
func (l *Logger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}
