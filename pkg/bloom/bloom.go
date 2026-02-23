// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package bloom wraps bits-and-blooms/bloom to provide a thread-safe
// Bloom filter for the list service's L1 cache layer.
// False-positive rate is ~0.1% at the configured capacity.
package bloom

import (
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
)

const (
	// DefaultCapacity is the expected number of elements in the filter.
	DefaultCapacity = 10_000_000
	// DefaultFPRate is the target false-positive rate (0.1%).
	DefaultFPRate = 0.001
)

// Filter is a thread-safe Bloom filter.
type Filter struct {
	mu sync.RWMutex
	bf *bloom.BloomFilter
}

// New returns a Filter sized for capacity elements at the given false-positive rate.
func New(capacity uint, fpRate float64) *Filter {
	return &Filter{bf: bloom.NewWithEstimates(capacity, fpRate)}
}

// NewDefault returns a Filter with DefaultCapacity and DefaultFPRate.
func NewDefault() *Filter {
	return New(DefaultCapacity, DefaultFPRate)
}

// Add inserts data into the filter.
func (f *Filter) Add(data []byte) {
	f.mu.Lock()
	f.bf.Add(data)
	f.mu.Unlock()
}

// Test reports whether data is possibly in the set.
// A false result guarantees the element is NOT present.
// A true result means the element is PROBABLY present (may be false positive).
func (f *Filter) Test(data []byte) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.bf.Test(data)
}

// Reset clears all entries from the filter.
func (f *Filter) Reset() {
	f.mu.Lock()
	f.bf.ClearAll()
	f.mu.Unlock()
}

// ApproxCount returns the estimated number of elements added.
func (f *Filter) ApproxCount() uint32 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.bf.ApproximatedSize()
}
