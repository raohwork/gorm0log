// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package gorm0log

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// Ignore denotes specified message is ignored.
func Ignore(l zerolog.Logger) *zerolog.Event { return l.Trace().Discard() }

// UseFatal denotes specified message belongs to Fatal level.
func UseFatal(l zerolog.Logger) *zerolog.Event { return l.Fatal() }

// UseError denotes specified message belongs to Error level.
func UseError(l zerolog.Logger) *zerolog.Event { return l.Error() }

// UseWarn denotes specified message belongs to Warn level.
func UseWarn(l zerolog.Logger) *zerolog.Event { return l.Warn() }

// UseInfo denotes specified message belongs to Info level.
func UseInfo(l zerolog.Logger) *zerolog.Event { return l.Info() }

// UseDebug denotes specified message belongs to Debug level.
func UseDebug(l zerolog.Logger) *zerolog.Event { return l.Debug() }

// UseTrace denotes specified message belongs to Trace level.
func UseTrace(l zerolog.Logger) *zerolog.Event { return l.Trace() }

// Config is a switch for extra features of Logger.
//
// Default value is fairly enough for general use:
//
//   - log general error as Error level
//   - record not found error as Debug level
//   - dump every sql as Debug level
//   - slow log message as Warn level
//   - disable slow log, as you should have your own slow threshold
//   - common json key
//   - ignore context values
//
// By default, [gorm.DB.Debug] shows sql dump. This can be changed by setting
// DumpLevel to [UseTrace] in [Config].
type Config struct {
	// Duration threshold of slow log, 0 or less disables it.
	SlowThreshold time.Duration
	// Log level of slow sql messages, default to [UseWarn].
	SlowLevel func(zerolog.Logger) *zerolog.Event
	// Key used to show time tracking info, default to "duration"
	Duration string

	// Log level for special error, default to log every error at Error level.
	// You might use it to change log level of non-critical errors like
	// [gorm.ErrRecordNotFound] or [gorm.ErrDuplicatedKey]. Helpers are
	// provided, see [IgnoreCommonErr] and [DebugCommonErr].
	ErrorLevel func(error, zerolog.Logger) *zerolog.Event

	// Do not log value of parameters.
	ParameterizedQueries bool

	// Dump SQL
	// Log level of sql dumping messages, default to [UseDebug].
	DumpLevel func(zerolog.Logger) *zerolog.Event
	// Adds execution time info to sql dumping message.
	DumpWithDuration bool
	// Key used to show sql dump, default to "sql".
	SQL string
	// Key used to show affected rows, default to "affected_rows".
	AffectedRows string

	// A function to log extra info, context value or call stacks for example.
	// This function is called only if the message is visible.
	Customize func(context.Context, *zerolog.Event)
}

func key(val, defaults string) string {
	if val != "" {
		return val
	}
	return defaults
}

// json key to store sql dump
func (c *Config) sqlKey() string { return key(c.SQL, "sql") }

// json key to store duration
func (c *Config) durKey() string { return key(c.Duration, "duration") }

// json key to store affected rows
func (c *Config) rowKey() string { return key(c.AffectedRows, "affected_rows") }

// log level of record not found message
func (c *Config) errLevel(err error, l zerolog.Logger) *zerolog.Event {
	if c.ErrorLevel == nil {
		return UseError(l).Err(err)
	}
	return c.ErrorLevel(err, l)
}

func level(val, defaults func(zerolog.Logger) *zerolog.Event) func(zerolog.Logger) *zerolog.Event {
	if val == nil {
		return defaults
	}
	return val
}

// log level of slow log message
func (c *Config) slowLevel(l zerolog.Logger) *zerolog.Event {
	return level(c.SlowLevel, UseWarn)(l)
}

// log level of sql dumping message
func (c *Config) dumpLevel(l zerolog.Logger) *zerolog.Event {
	return level(c.DumpLevel, UseDebug)(l)
}

// calls cutsomizing function
func (c *Config) custom(ctx context.Context) func(*zerolog.Event) {
	return func(ev *zerolog.Event) {
		if c.Customize == nil {
			return
		}
		c.Customize(ctx, ev)
	}
}

// format of error log message
func (c *Config) logErr(err error, f func() (string, int64)) func(*zerolog.Event) {
	return func(ev *zerolog.Event) {
		sql, rows := f()
		ev.Err(err).Str(c.sqlKey(), sql)
		if rows != -1 {
			ev.Int64(c.rowKey(), rows)
		}
	}
}

// format of slow log message
func (c *Config) logSlow(dur time.Duration, f func() (string, int64)) func(*zerolog.Event) {
	return func(ev *zerolog.Event) {
		sql, rows := f()
		ev.Dur(c.durKey(), dur).Str(c.sqlKey(), sql)
		if rows != -1 {
			ev.Int64(c.rowKey(), rows)
		}
	}
}

// format of sql dumping message
func (c *Config) logDump(dur time.Duration, f func() (string, int64)) func(*zerolog.Event) {
	return func(ev *zerolog.Event) {
		if c.DumpWithDuration {
			ev.Dur(c.durKey(), dur)
		}

		sql, rows := f()
		ev.Str(c.sqlKey(), sql)
		if rows != -1 {
			ev.Int64(c.rowKey(), rows)
		}
	}
}
