// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"
	"time"

	"github.com/yourorg/riskengine/pkg/dsl"
)

// commonFormats are tried in order when parsing a time string.
var commonFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
	"20060102",
}

func parseTimeStr(s string) (time.Time, error) {
	for _, f := range commonFormats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time %q", s)
}

func registerTime(reg *dsl.FunctionRegistry) error {
	fns := []struct {
		name string
		impl dsl.FuncImpl
	}{
		// now() → int  (Unix timestamp in seconds)
		{"now", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			return dsl.IntValue(time.Now().Unix()), nil
		}},
		// nowMs() → int  (Unix timestamp in milliseconds)
		{"nowMs", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			return dsl.IntValue(time.Now().UnixMilli()), nil
		}},
		// daysSince(timeStr) → int  (number of full days since the given date)
		{"daysSince", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 1 {
				return dsl.NilValue(), fmt.Errorf("daysSince() requires 1 arg")
			}
			t, err := parseTimeStr(args[0].String())
			if err != nil {
				return dsl.IntValue(0), nil // graceful degradation
			}
			days := int64(time.Since(t).Hours() / 24)
			return dsl.IntValue(days), nil
		}},
		// hoursSince(timeStr) → int
		{"hoursSince", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 1 {
				return dsl.NilValue(), fmt.Errorf("hoursSince() requires 1 arg")
			}
			t, err := parseTimeStr(args[0].String())
			if err != nil {
				return dsl.IntValue(0), nil
			}
			return dsl.IntValue(int64(time.Since(t).Hours())), nil
		}},
		// secondsSince(timeStr) → int
		{"secondsSince", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 1 {
				return dsl.NilValue(), fmt.Errorf("secondsSince() requires 1 arg")
			}
			t, err := parseTimeStr(args[0].String())
			if err != nil {
				return dsl.IntValue(0), nil
			}
			return dsl.IntValue(int64(time.Since(t).Seconds())), nil
		}},
		// toUnix(timeStr) → int  (parse and return Unix timestamp)
		{"toUnix", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 1 {
				return dsl.NilValue(), fmt.Errorf("toUnix() requires 1 arg")
			}
			t, err := parseTimeStr(args[0].String())
			if err != nil {
				return dsl.IntValue(0), nil
			}
			return dsl.IntValue(t.Unix()), nil
		}},
		// hour() → int  (current UTC hour, 0-23)
		{"hour", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			return dsl.IntValue(int64(time.Now().UTC().Hour())), nil
		}},
		// weekday() → int  (0=Sunday … 6=Saturday)
		{"weekday", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			return dsl.IntValue(int64(time.Now().UTC().Weekday())), nil
		}},
	}

	for _, f := range fns {
		if err := reg.RegisterFunc(f.name, f.impl); err != nil {
			return fmt.Errorf("builtins.time: %w", err)
		}
	}
	return nil
}
