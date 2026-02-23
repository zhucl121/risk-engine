// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package list provides a tiered list lookup service.
//
// Lookup order:
//   L1 – in-process Bloom filter  (~0.1ms, false-positive 0.1%)
//   L2 – Redis                    (~2–5ms, exact match)
//   L3 – persistent store          (~5–15ms, exact match)
//
// A miss at L1 guarantees the entity is not in any list (no false negatives).
// An L1 hit triggers an L2 lookup for exact status resolution.
package list

import (
	"context"
	"errors"
	"time"
)

// Status represents the list membership of an entity.
type Status int

const (
	StatusNotFound  Status = iota // entity not found in any list
	StatusBlacklist               // entity is blacklisted
	StatusWhitelist               // entity is whitelisted
	StatusGraylist                // entity is under observation
)

// Query identifies an entity for list lookup.
type Query struct {
	// Kind categorises the entity: "user", "device", "ip", "card", etc.
	Kind  string
	Value string
}

// Service is the list lookup interface.
// All methods must be safe for concurrent use.
type Service interface {
	// Check returns the list status for the entity described by q.
	Check(ctx context.Context, q *Query) (Status, error)

	// Add inserts or updates an entity's list status with an optional TTL.
	// A zero TTL means no expiry.
	Add(ctx context.Context, q *Query, status Status, ttl time.Duration) error

	// Remove deletes an entity from all lists.
	Remove(ctx context.Context, q *Query) error
}

// ErrNotConfigured is returned when the list service has no storage backend.
var ErrNotConfigured = errors.New("list service not configured")
