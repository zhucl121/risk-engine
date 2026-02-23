// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RuleRepository is the persistence interface for risk rules.
// All implementations must be safe for concurrent use.
type RuleRepository interface {
	// ListActive returns all enabled rules for the given scene code.
	// An empty sceneCode returns rules for all scenes.
	ListActive(ctx context.Context, sceneCode string) ([]*RuleRecord, error)

	// ListUpdatedSince returns rules whose UpdatedAt is after the given time.
	// Used by the hot-reload watcher for incremental updates.
	ListUpdatedSince(ctx context.Context, since time.Time) ([]*RuleRecord, error)

	// GetByID returns a single rule by its numeric ID.
	GetByID(ctx context.Context, id int64) (*RuleRecord, error)

	// Create inserts a new rule and returns its auto-assigned ID.
	Create(ctx context.Context, r *RuleRecord) (int64, error)

	// Update saves changes to an existing rule.
	// Implementations MUST enforce optimistic locking on the Version field.
	Update(ctx context.Context, r *RuleRecord) error

	// SoftDelete marks a rule as disabled (status=0) without deleting the row.
	SoftDelete(ctx context.Context, id int64) error
}

// ─── In-memory fake (used in unit tests) ─────────────────────────────────────

// FakeRepository is an in-memory RuleRepository implementation for tests.
// It is NOT safe for production use (no persistence, no ordering guarantees).
type FakeRepository struct {
	mu      sync.RWMutex
	records map[int64]*RuleRecord
	nextID  int64
}

// NewFakeRepository returns an empty FakeRepository.
func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		records: make(map[int64]*RuleRecord),
		nextID:  1,
	}
}

func (f *FakeRepository) ListActive(ctx context.Context, sceneCode string) ([]*RuleRecord, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var out []*RuleRecord
	for _, r := range f.records {
		if r.Status != 1 {
			continue
		}
		if sceneCode != "" && r.SceneCode != sceneCode {
			continue
		}
		cp := *r
		out = append(out, &cp)
	}
	return out, nil
}

func (f *FakeRepository) ListUpdatedSince(ctx context.Context, since time.Time) ([]*RuleRecord, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var out []*RuleRecord
	for _, r := range f.records {
		if r.UpdatedAt.After(since) {
			cp := *r
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (f *FakeRepository) GetByID(ctx context.Context, id int64) (*RuleRecord, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	r, ok := f.records[id]
	if !ok {
		return nil, fmt.Errorf("rule %d not found", id)
	}
	cp := *r
	return &cp, nil
}

func (f *FakeRepository) Create(ctx context.Context, r *RuleRecord) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id := f.nextID
	f.nextID++
	cp := *r
	cp.ID = id
	cp.Version = 1
	cp.CreatedAt = time.Now()
	cp.UpdatedAt = time.Now()
	f.records[id] = &cp
	return id, nil
}

func (f *FakeRepository) Update(ctx context.Context, r *RuleRecord) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	existing, ok := f.records[r.ID]
	if !ok {
		return fmt.Errorf("rule %d not found", r.ID)
	}
	if existing.Version != r.Version {
		return fmt.Errorf("optimistic lock conflict: rule %d version %d != %d", r.ID, existing.Version, r.Version)
	}
	cp := *r
	cp.Version = r.Version + 1
	cp.UpdatedAt = time.Now()
	f.records[r.ID] = &cp
	return nil
}

func (f *FakeRepository) SoftDelete(ctx context.Context, id int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.records[id]
	if !ok {
		return fmt.Errorf("rule %d not found", id)
	}
	r.Status = 0
	r.UpdatedAt = time.Now()
	return nil
}
