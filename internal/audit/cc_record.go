// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package audit

import "time"

// ChampionChallengerRecord captures the outcome of one challenger execution
// side-by-side with the champion's result.
//
// Records are written asynchronously under the "cc_audit" key so data
// scientists can run offline statistical analysis (win rate, error rate,
// latency delta) without impacting production performance.
//
// Key field layout mirrors ShadowRecord for tooling consistency, but adds
// ExperimentID and ChallengerID to enable per-variant querying.
type ChampionChallengerRecord struct {
	// RequestID matches the production request, enabling join with the main audit table.
	RequestID string `json:"request_id"`
	// SceneCode is the production scene.
	SceneCode string `json:"scene_code"`
	// ExperimentID groups all variants of an experiment for analysis.
	ExperimentID string `json:"experiment_id"`
	// ChallengerID identifies which challenger variant produced this record.
	ChallengerID string `json:"challenger_id"`

	// User / device identifiers (masked for privacy).
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`

	// Champion outcome (the decision that was actually returned to the caller).
	ChampionDecision  string `json:"champion_decision"`
	ChampionRiskScore int    `json:"champion_risk_score"`
	ChampionCostMs    int64  `json:"champion_cost_ms"`
	ChampionHitRules  []string           `json:"champion_hit_rules,omitempty"`
	ChampionModels    map[string]float64 `json:"champion_models,omitempty"`

	// Challenger outcome (executed concurrently; NOT returned to the caller).
	ChallengerDecision  string             `json:"challenger_decision"`
	ChallengerRiskScore int                `json:"challenger_risk_score"`
	ChallengerCostMs    int64              `json:"challenger_cost_ms"`
	ChallengerHitRules  []string           `json:"challenger_hit_rules,omitempty"`
	ChallengerModels    map[string]float64 `json:"challenger_models,omitempty"`
	ChallengerReasons   []string           `json:"challenger_reasons,omitempty"`

	// Agreement indicates whether champion and challenger reached the same decision.
	// Useful for quick dashboard metrics (agreement_rate, flip_rate).
	Agreement bool `json:"agreement"`

	CreatedAt time.Time `json:"created_at"`
}

// NewChampionChallengerRecord constructs a ChampionChallengerRecord.
func NewChampionChallengerRecord(
	requestID, sceneCode, experimentID, challengerID,
	userID, deviceID string,
	champDecision string, champScore int, champCostMs int64,
	champHitRules []string, champModels map[string]float64,
	challDecision string, challScore int, challCostMs int64,
	challHitRules []string, challModels map[string]float64,
	challReasons []string,
) *ChampionChallengerRecord {
	return &ChampionChallengerRecord{
		RequestID:           requestID,
		SceneCode:           sceneCode,
		ExperimentID:        experimentID,
		ChallengerID:        challengerID,
		UserID:              maskUserID(userID),
		DeviceID:            deviceID,
		ChampionDecision:    champDecision,
		ChampionRiskScore:   champScore,
		ChampionCostMs:      champCostMs,
		ChampionHitRules:    champHitRules,
		ChampionModels:      champModels,
		ChallengerDecision:  challDecision,
		ChallengerRiskScore: challScore,
		ChallengerCostMs:    challCostMs,
		ChallengerHitRules:  challHitRules,
		ChallengerModels:    challModels,
		ChallengerReasons:   challReasons,
		Agreement:           champDecision == challDecision,
		CreatedAt:           time.Now(),
	}
}
