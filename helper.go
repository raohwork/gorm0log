// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package gorm0log

import (
	"context"
	"errors"
	"runtime"
	"strings"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// LogSource creates a function to provide caller info.
//
// It iterates the stack to find fist file that matches any of keywords, puts
// filename and line number to json field "source_file" and "source_line".
func LogSource(keywords ...string) func(context.Context, *zerolog.Event) {
	return func(_ context.Context, ev *zerolog.Event) {
		fn, line, ok := "", 0, true
		cnt := 2
		for ok {
			_, fn, line, ok = runtime.Caller(cnt)
			cnt++
			if !ok {
				return
			}

			for _, kw := range keywords {
				if strings.Contains(fn, kw) {
					ev.Str("source_file", fn)
					ev.Int("source_line", line)
					return
				}
			}
		}
	}
}

// LogErrorAt creates a function to be used at ErrorLevel of [Config]. It compares
// error using cmpErr, use specified level to log it if matched, Error level
// otherwise.
func LogErrorAt(level func(zerolog.Logger) *zerolog.Event, cmpErr func(error) bool) func(error, zerolog.Logger) *zerolog.Event {
	return func(err error, l zerolog.Logger) *zerolog.Event {
		if cmpErr(err) {
			return level(l)
		}
		return UseError(l)
	}
}

// CommonError detects if err is [gorm.ErrDuplicatedKey] or [gorm.ErrRecordNotFound].
func CommonError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, gorm.ErrDuplicatedKey)
}

// IgnoreCommonErr is shortcut of LogErrorAt(UseTrace, CommonError).
func IgnoreCommonErr(e error, l zerolog.Logger) *zerolog.Event {
	return LogErrorAt(UseTrace, CommonError)(e, l)
}

// DebugCommonErr is shortcut of LogErrorAt(UseDebug, CommonError).
func DebugCommonErr(e error, l zerolog.Logger) *zerolog.Event {
	return LogErrorAt(UseDebug, CommonError)(e, l)
}
