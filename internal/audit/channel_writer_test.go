// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package audit_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/yourorg/riskengine/internal/audit"
	"github.com/yourorg/riskengine/internal/engine"
)

func newRecord(id string) *audit.Record {
	return &audit.Record{
		RequestID: id,
		SceneCode: "test_scene",
		Decision:  engine.DecisionPass,
		RiskScore: 10,
		CostMs:    5,
		CreatedAt: time.Now(),
	}
}

func TestChannelWriter_WriteAndFlush(t *testing.T) {
	logger := zaptest.NewLogger(t)
	w := audit.NewChannelWriter(logger, 32)
	defer func() { require.NoError(t, w.Close()) }()

	ctx := context.Background()
	for i := range 10 {
		require.NoError(t, w.Write(ctx, newRecord(string(rune('a'+i)))))
	}

	flushCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	require.NoError(t, w.Flush(flushCtx))
	assert.Equal(t, int64(0), w.DroppedCount())
}

func TestChannelWriter_DropsWhenFull(t *testing.T) {
	logger := zap.NewNop()
	// Buffer of 1 so subsequent writes are dropped.
	w := audit.NewChannelWriter(logger, 1)

	ctx := context.Background()
	// Pause the consumer by filling the buffer before it drains.
	for range 50 {
		_ = w.Write(ctx, newRecord("x"))
	}

	// At least some must be dropped (buffer=1, 50 writes).
	// Give consumer time to drain then check.
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, w.Close())

	assert.Positive(t, w.DroppedCount())
}

func TestChannelWriter_CloseIdempotent(t *testing.T) {
	logger := zap.NewNop()
	w := audit.NewChannelWriter(logger, 8)
	require.NoError(t, w.Close())
	require.NoError(t, w.Close()) // second Close must not panic
}

func TestChannelWriter_WriteAfterClose(t *testing.T) {
	logger := zap.NewNop()
	w := audit.NewChannelWriter(logger, 8)
	require.NoError(t, w.Close())

	// Write after close must be a no-op (no panic, no error).
	err := w.Write(context.Background(), newRecord("after-close"))
	require.NoError(t, err)
}
