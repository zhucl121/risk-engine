// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
)

const defaultBufferSize = 4096

// ChannelWriter is a non-blocking audit Writer that enqueues records into a
// buffered channel and drains them in a background goroutine.
//
// In production this would forward records to Kafka; here we serialise to
// JSON and emit a structured log entry so the system is self-contained.
//
// Design guarantees:
//   - Write() never blocks the caller; full channel → record dropped + counter
//   - Flush() blocks until the channel is drained or ctx is cancelled
//   - Close() signals the consumer to stop; calling Write after Close is a no-op
type ChannelWriter struct {
	ch       chan *Record
	logger   *zap.Logger
	wg       sync.WaitGroup
	dropped  atomic.Int64
	once     sync.Once // guards Close
	closed   atomic.Bool
}

// NewChannelWriter creates a ChannelWriter and starts the background consumer.
// bufferSize is the internal channel capacity; pass 0 to use the default (4096).
func NewChannelWriter(logger *zap.Logger, bufferSize int) *ChannelWriter {
	if bufferSize <= 0 {
		bufferSize = defaultBufferSize
	}
	w := &ChannelWriter{
		ch:     make(chan *Record, bufferSize),
		logger: logger,
	}
	w.wg.Add(1)
	go w.consume()
	return w
}

// Write enqueues r for asynchronous delivery.
// It returns immediately; if the channel is full the record is silently dropped
// and the drop counter is incremented.
func (w *ChannelWriter) Write(_ context.Context, r *Record) error {
	if w.closed.Load() {
		return nil
	}
	select {
	case w.ch <- r:
	default:
		n := w.dropped.Add(1)
		w.logger.Warn("audit channel full, record dropped",
			zap.Int64("total_dropped", n),
			zap.String("request_id", r.RequestID),
		)
	}
	return nil
}

// Flush blocks until the internal buffer is empty or ctx is cancelled.
func (w *ChannelWriter) Flush(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		for len(w.ch) > 0 {
			select {
			case <-ctx.Done():
				close(done)
				return
			default:
			}
		}
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("audit.Flush: %w", ctx.Err())
	}
}

// Close signals the background consumer to stop after draining the channel.
// It blocks until the consumer goroutine exits.
func (w *ChannelWriter) Close() error {
	w.once.Do(func() {
		w.closed.Store(true)
		close(w.ch)
	})
	w.wg.Wait()
	return nil
}

// DroppedCount returns the total number of records dropped due to a full channel.
func (w *ChannelWriter) DroppedCount() int64 {
	return w.dropped.Load()
}

// consume is the background goroutine that reads from the channel and persists records.
func (w *ChannelWriter) consume() {
	defer w.wg.Done()
	for r := range w.ch {
		w.emit(r)
	}
}

// emit serialises the record to JSON and writes a structured log entry.
// In a production Kafka integration this would produce to risk.decisions topic.
func (w *ChannelWriter) emit(r *Record) {
	b, err := json.Marshal(r)
	if err != nil {
		w.logger.Error("audit: marshal failed",
			zap.String("request_id", r.RequestID),
			zap.Error(err),
		)
		return
	}
	w.logger.Info("audit",
		zap.String("request_id", r.RequestID),
		zap.String("scene_code", r.SceneCode),
		zap.String("decision", string(r.Decision)),
		zap.Int("risk_score", r.RiskScore),
		zap.Int64("cost_ms", r.CostMs),
		zap.ByteString("payload", b),
	)
}
