// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package audit writes immutable decision records to Kafka for compliance,
// forensics, and model feedback loops. PII fields are masked before writing.
package audit

import (
	"context"
	"time"
)

// Record is an immutable snapshot of a single risk decision.
// Sensitive fields (user ID, card number) are masked before persistence.
// All engine/feature/rule types are replaced by primitive equivalents to
// prevent import cycles (audit ← engine ← orchestrator ← audit).
type Record struct {
	RequestID string
	SceneCode string
	// UserID is masked: only last 4 characters visible.
	UserID   string
	DeviceID string
	// Decision holds the decision string (e.g. "PASS", "REJECT").
	Decision    string
	RiskScore   int
	// Features is the masked feature map serialised as map[string]any.
	Features    map[string]any
	ModelScores map[string]float64
	HitRules    []string
	RiskReasons []string
	CostMs      int64
	CreatedAt   time.Time
}

// Writer persists audit records asynchronously.
// Implementations must guarantee at-least-once delivery.
type Writer interface {
	// Write enqueues a record for asynchronous delivery.
	// It returns an error only if the record cannot be enqueued locally.
	Write(ctx context.Context, r *Record) error

	// Flush blocks until all buffered records are delivered or ctx expires.
	Flush(ctx context.Context) error

	// Close releases all resources. Flush should be called first.
	Close() error
}

// maskUserID returns the user ID with all but the last 4 characters replaced by '*'.
func maskUserID(id string) string {
	if len(id) <= 4 {
		return id
	}
	masked := make([]byte, len(id))
	for i := range len(id) - 4 {
		masked[i] = '*'
	}
	copy(masked[len(id)-4:], id[len(id)-4:])
	return string(masked)
}

// NewRecord builds an audit Record, masking PII in the process.
// All caller-side types are passed as primitives to avoid import cycles.
func NewRecord(
	requestID, sceneCode, userID, deviceID, decision string,
	riskScore int,
	features map[string]any,
	modelScores map[string]float64,
	hitRules, riskReasons []string,
	costMs int64,
) *Record {
	return &Record{
		RequestID:   requestID,
		SceneCode:   sceneCode,
		UserID:      maskUserID(userID),
		DeviceID:    deviceID,
		Decision:    decision,
		RiskScore:   riskScore,
		Features:    maskPIIMap(features),
		ModelScores: modelScores,
		HitRules:    hitRules,
		RiskReasons: riskReasons,
		CostMs:      costMs,
		CreatedAt:   time.Now(),
	}
}

// maskPIIMap returns a copy of the feature map with known PII keys masked.
func maskPIIMap(m map[string]any) map[string]any {
	piiKeys := map[string]bool{
		"user.phone":       true,
		"user.id_number":   true,
		"payment.card_num": true,
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		if piiKeys[k] {
			out[k] = "****"
			continue
		}
		out[k] = v
	}
	return out
}
