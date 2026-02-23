// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package rule

import "time"

// RuleRecord is the database-level representation of a risk rule.
// It maps 1:1 to the risk_rules table.
type RuleRecord struct {
	ID             int64
	RuleKey        string // globally unique, e.g. "DEVICE_MULTI_ACCOUNT_001"
	Name           string
	GroupName      string
	SceneCode      string
	Priority       int
	ConditionDSL   string // DSL string compiled by pkg/dsl
	ConditionAST   []byte // JSON blob for visual rule builder (nullable)
	ActionDecision string // REJECT | MANUAL_REVIEW | PASS
	ActionRiskCode string
	ActionScore    int
	Status         int8 // 1=enabled, 0=disabled
	Version        int  // optimistic lock
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
