// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package scene

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// cacheEntry holds a scene's specs together with the last-fetched wall clock
// time (used by callers that need TTL-aware invalidation).
type cacheEntry struct {
	specs     []ExtraParamSpec
	fetchedAt time.Time
}

// ExtraParamLoader is a thread-safe, hot-reloading cache of ExtraParamSpec
// records backed by an ExtraParamRepository.
//
//   - First call for a scene triggers a synchronous DB fetch that is stored
//     in the internal sync.Map.
//   - A background goroutine (started via StartWatcher) periodically calls
//     ListUpdatedSince to incrementally update changed records, keeping the
//     cache warm without full-table scans.
type ExtraParamLoader struct {
	repo     ExtraParamRepository
	cache    sync.Map      // key: sceneCode (string) → *cacheEntry
	interval time.Duration // background watcher poll period
	logger   *zap.Logger
}

// NewExtraParamLoader constructs an ExtraParamLoader.
// interval controls how often the background watcher refreshes the cache;
// 0 defaults to 30 seconds.
func NewExtraParamLoader(repo ExtraParamRepository, interval time.Duration, logger *zap.Logger) *ExtraParamLoader {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ExtraParamLoader{
		repo:     repo,
		interval: interval,
		logger:   logger.Named("scene.extra_loader"),
	}
}

// Load returns all active ExtraParamSpecs for the given scene.
// On first call it fetches from the DB; subsequent calls are served from
// the in-memory cache.
func (l *ExtraParamLoader) Load(ctx context.Context, sceneCode string) ([]ExtraParamSpec, error) {
	if v, ok := l.cache.Load(sceneCode); ok {
		return v.(*cacheEntry).specs, nil
	}
	return l.refresh(ctx, sceneCode)
}

// refresh fetches fresh data from the repository and updates the cache.
func (l *ExtraParamLoader) refresh(ctx context.Context, sceneCode string) ([]ExtraParamSpec, error) {
	ptrs, err := l.repo.ListByScene(ctx, sceneCode)
	if err != nil {
		return nil, err
	}
	specs := make([]ExtraParamSpec, 0, len(ptrs))
	for _, p := range ptrs {
		specs = append(specs, *p)
	}
	l.cache.Store(sceneCode, &cacheEntry{specs: specs, fetchedAt: time.Now()})
	return specs, nil
}

// Invalidate removes the cached entry for the scene, forcing the next Load
// call to re-fetch from the database.
func (l *ExtraParamLoader) Invalidate(sceneCode string) {
	l.cache.Delete(sceneCode)
}

// StartWatcher launches a background goroutine that polls ListUpdatedSince
// every l.interval and applies incremental updates to the cache.
// The goroutine exits when ctx is cancelled.
func (l *ExtraParamLoader) StartWatcher(ctx context.Context) {
	go l.watch(ctx)
}

func (l *ExtraParamLoader) watch(ctx context.Context) {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	lastCheck := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			since := lastCheck
			lastCheck = t

			updated, err := l.repo.ListUpdatedSince(ctx, since)
			if err != nil {
				l.logger.Warn("hot-reload ListUpdatedSince failed", zap.Error(err))
				continue
			}
			if len(updated) == 0 {
				continue
			}

			// Group changed specs by scene and merge into the cache.
			byScene := make(map[string][]*ExtraParamSpec)
			for _, s := range updated {
				byScene[s.SceneCode] = append(byScene[s.SceneCode], s)
			}
			for sceneCode := range byScene {
				// Full re-fetch for simplicity; the ListUpdatedSince result
				// may contain partial data, and a full refresh is cheaper
				// than maintaining per-key delta logic.
				if _, err := l.refresh(ctx, sceneCode); err != nil {
					l.logger.Warn("hot-reload refresh failed",
						zap.String("scene", sceneCode),
						zap.Error(err),
					)
				} else {
					l.logger.Debug("hot-reload ok", zap.String("scene", sceneCode))
				}
			}
		}
	}
}
