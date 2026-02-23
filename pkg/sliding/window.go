// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package sliding provides a Redis-backed sliding-window velocity counter.
// Each Count call atomically removes expired entries, adds the current event,
// and returns the count within the window — all in a single Lua script round-trip.
package sliding

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed lua/count.lua
var countLua string

// Window is a Redis sliding-window counter. It is safe for concurrent use.
type Window struct {
	rdb    redis.UniversalClient
	script *redis.Script
}

// New returns a new Window backed by the provided Redis client.
func New(rdb redis.UniversalClient) *Window {
	return &Window{
		rdb:    rdb,
		script: redis.NewScript(countLua),
	}
}

// Count increments the counter for key and returns the number of events
// within [now-window, now]. The key expires automatically after the window
// duration plus a one-second buffer.
func (w *Window) Count(ctx context.Context, key string, window time.Duration) (int64, error) {
	now := time.Now().UnixMilli()
	min := now - window.Milliseconds()
	ttlSec := int(window.Seconds()) + 1

	result, err := w.script.Run(ctx, w.rdb, []string{key}, min, now, now, ttlSec).Int64()
	if err != nil {
		return 0, fmt.Errorf("sliding window count %q: %w", key, err)
	}
	return result, nil
}

// Get returns the current count within [now-window, now] without incrementing.
func (w *Window) Get(ctx context.Context, key string, window time.Duration) (int64, error) {
	now := time.Now().UnixMilli()
	min := now - window.Milliseconds()

	result, err := w.rdb.ZCount(ctx, key, fmt.Sprintf("%d", min), fmt.Sprintf("%d", now)).Result()
	if err != nil {
		return 0, fmt.Errorf("sliding window get %q: %w", key, err)
	}
	return result, nil
}
