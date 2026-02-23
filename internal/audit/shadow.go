// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package audit

import "time"

// ShadowRecord captures the outcome of a shadow (dry-run) policy execution.
//
// A shadow policy runs alongside the real pipeline but its decision is NEVER
// returned to the caller. ShadowRecords are written asynchronously to the
// audit log under the "shadow_audit" key so analysts can compare shadow vs.
// production decisions offline.
//
// Fields mirror Record to keep analysis tooling uniform.
type ShadowRecord struct {
	// RequestID is the same ID as the real request, enabling join queries.
	RequestID string `json:"request_id"`
	// SceneCode of the real (production) scene.
	SceneCode string `json:"scene_code"`
	// ShadowSceneCode identifies the policy that ran in shadow mode.
	ShadowSceneCode string `json:"shadow_scene_code"`
	// ShadowVersion is the version tag of the shadow policy (for traceability).
	ShadowVersion string `json:"shadow_version"`

	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`

	// Decision and RiskScore from the shadow execution.
	Decision  string `json:"decision"`
	RiskScore int    `json:"risk_score"`

	// ProductionDecision is the real decision returned to the caller.
	// Analysts can compare Decision vs. ProductionDecision to evaluate policy changes.
	ProductionDecision string `json:"production_decision"`
	ProductionScore    int    `json:"production_score"`

	HitRules    []string           `json:"hit_rules,omitempty"`
	ModelScores map[string]float64 `json:"model_scores,omitempty"`
	RiskReasons []string           `json:"risk_reasons,omitempty"`
	CostMs      int64              `json:"cost_ms"`
	CreatedAt   time.Time          `json:"created_at"`
}

// NewShadowRecord builds a ShadowRecord from shadow and production results.
func NewShadowRecord(
	requestID, realScene, shadowScene, shadowVersion,
	userID, deviceID string,
	shadowDecision string, shadowScore int,
	prodDecision string, prodScore int,
	hitRules []string,
	modelScores map[string]float64,
	riskReasons []string,
	costMs int64,
) *ShadowRecord {
	return &ShadowRecord{
		RequestID:          requestID,
		SceneCode:          realScene,
		ShadowSceneCode:    shadowScene,
		ShadowVersion:      shadowVersion,
		UserID:             maskUserID(userID),
		DeviceID:           deviceID,
		Decision:           shadowDecision,
		RiskScore:          shadowScore,
		ProductionDecision: prodDecision,
		ProductionScore:    prodScore,
		HitRules:           hitRules,
		ModelScores:        modelScores,
		RiskReasons:        riskReasons,
		CostMs:             costMs,
		CreatedAt:          time.Now(),
	}
}
