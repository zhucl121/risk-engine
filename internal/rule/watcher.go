// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// DBWatcher polls the database for rule changes and triggers a hot-reload
// on the provided Evaluator via Reload().
//
// The watcher runs in a background goroutine started by Start().
// Stop the watcher by cancelling the context passed to Start().
type DBWatcher struct {
	loader   *DBLoader
	evaluator Evaluator
	interval  time.Duration
	logger    *zap.Logger
}

// NewDBWatcher creates a watcher that polls every interval.
// interval should be <= 30s to meet the hot-reload propagation SLA.
func NewDBWatcher(loader *DBLoader, evaluator Evaluator, interval time.Duration, logger *zap.Logger) *DBWatcher {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &DBWatcher{
		loader:   loader,
		evaluator: evaluator,
		interval:  interval,
		logger:   logger,
	}
}

// Start begins the polling loop. It blocks until ctx is cancelled.
// Run it in a goroutine: go watcher.Start(ctx).
func (w *DBWatcher) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	lastCheck := time.Now()
	w.logger.Info("DBWatcher started", zap.Duration("interval", w.interval))

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("DBWatcher stopped")
			return
		case tick := <-ticker.C:
			w.poll(ctx, lastCheck)
			lastCheck = tick
		}
	}
}

func (w *DBWatcher) poll(ctx context.Context, since time.Time) {
	updated, err := w.loader.LoadUpdatedSince(ctx, since)
	if err != nil {
		w.logger.Error("DBWatcher poll failed", zap.Error(err))
		return
	}
	if len(updated) == 0 {
		return
	}

	// Load full active rule set (incremental merge is complex; full reload is simpler
	// given typical rule counts of < 1000 and 30s interval).
	all, err := w.loader.LoadAll(ctx, "")
	if err != nil {
		w.logger.Error("DBWatcher full reload failed", zap.Error(err))
		return
	}
	if err := w.evaluator.Reload(all); err != nil {
		w.logger.Error("DBWatcher evaluator reload failed", zap.Error(err))
		return
	}
	w.logger.Info("DBWatcher hot-reload complete",
		zap.Int("updated_count", len(updated)),
		zap.Int("total_rules", len(all)),
	)
}

// ForceReload triggers an immediate full reload regardless of update timestamps.
// Called by the management API's POST /admin/v1/rules/reload endpoint.
func (w *DBWatcher) ForceReload(ctx context.Context) error {
	all, err := w.loader.LoadAll(ctx, "")
	if err != nil {
		return err
	}
	return w.evaluator.Reload(all)
}
