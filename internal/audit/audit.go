// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package audit writes immutable decision records to Kafka for compliance,
// forensics, and model feedback loops. PII fields are masked before writing.
package audit

import (
	"context"
	"time"

	"github.com/yourorg/riskengine/internal/engine"
	"github.com/yourorg/riskengine/internal/feature"
	"github.com/yourorg/riskengine/internal/rule"
)

// Record is an immutable snapshot of a single risk decision.
// Sensitive fields (user ID, card number) are masked before persistence.
type Record struct {
	RequestID   string
	SceneCode   string
	// UserID is masked: only last 4 characters visible.
	UserID      string
	DeviceID    string
	Decision    engine.Decision
	RiskScore   int
	// Features contains the full feature map with PII fields masked.
	Features    feature.Map
	RuleResults []*rule.Result
	ModelScores map[string]float64
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

// NewRecord builds an audit Record from a decision, masking PII in the process.
func NewRecord(req *engine.DecisionRequest, res *engine.DecisionResult, features feature.Map, ruleResults []*rule.Result) *Record {
	return &Record{
		RequestID:   req.RequestID,
		SceneCode:   req.SceneCode,
		UserID:      maskUserID(req.UserID),
		DeviceID:    req.DeviceID,
		Decision:    res.Decision,
		RiskScore:   res.RiskScore,
		Features:    maskPII(features),
		RuleResults: ruleResults,
		ModelScores: res.ModelScores,
		CostMs:      res.CostMs,
		CreatedAt:   time.Now(),
	}
}

// maskPII returns a copy of the feature map with known PII keys masked.
func maskPII(m feature.Map) feature.Map {
	piiKeys := map[string]bool{
		"user.phone":       true,
		"user.id_number":   true,
		"payment.card_num": true,
	}
	out := make(feature.Map, len(m))
	for k, v := range m {
		if piiKeys[k] {
			out[k] = feature.Value{Kind: feature.KindString, StrVal: "****"}
			continue
		}
		out[k] = v
	}
	return out
}
