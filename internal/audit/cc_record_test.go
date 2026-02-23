// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"testing"
)

func TestNewChampionChallengerRecord_Agreement(t *testing.T) {
	rec := NewChampionChallengerRecord(
		"req-1", "payment", "exp-1", "challenger-A",
		"user-1", "dev-1",
		"PASS", 200, 10,
		[]string{"rule_1"}, map[string]float64{"model_a": 0.2},
		"PASS", 300, 8,
		nil, nil,
		nil,
	)
	if !rec.Agreement {
		t.Error("expected agreement=true when both decisions are PASS")
	}
	if rec.UserID == "user-1" {
		t.Error("userID should be masked")
	}
}

func TestNewChampionChallengerRecord_Disagreement(t *testing.T) {
	rec := NewChampionChallengerRecord(
		"req-2", "payment", "exp-1", "challenger-B",
		"user-2", "dev-2",
		"PASS", 100, 5,
		nil, nil,
		"REJECT", 900, 12,
		nil, nil,
		[]string{"high_risk_indicator"},
	)
	if rec.Agreement {
		t.Error("expected agreement=false when champion=PASS, challenger=REJECT")
	}
	if rec.ChampionDecision != "PASS" {
		t.Errorf("got champion decision %q, want PASS", rec.ChampionDecision)
	}
	if rec.ChallengerDecision != "REJECT" {
		t.Errorf("got challenger decision %q, want REJECT", rec.ChallengerDecision)
	}
}

func TestNewChampionChallengerRecord_Fields(t *testing.T) {
	champModels := map[string]float64{"fraud_v1": 0.3}
	challModels := map[string]float64{"fraud_v3": 0.85}
	rec := NewChampionChallengerRecord(
		"req-3", "scene_A", "exp-2", "v3_eval",
		"u3", "d3",
		"MANUAL_REVIEW", 600, 15,
		[]string{"r1", "r2"}, champModels,
		"REJECT", 950, 9,
		[]string{"r_new"}, challModels,
		[]string{"reason_x"},
	)
	if rec.ExperimentID != "exp-2" {
		t.Errorf("ExperimentID mismatch: got %q", rec.ExperimentID)
	}
	if rec.ChallengerID != "v3_eval" {
		t.Errorf("ChallengerID mismatch: got %q", rec.ChallengerID)
	}
	if rec.ChampionRiskScore != 600 {
		t.Errorf("ChampionRiskScore mismatch: got %d", rec.ChampionRiskScore)
	}
	if rec.ChallengerRiskScore != 950 {
		t.Errorf("ChallengerRiskScore mismatch: got %d", rec.ChallengerRiskScore)
	}
	if len(rec.ChallengerReasons) != 1 || rec.ChallengerReasons[0] != "reason_x" {
		t.Errorf("ChallengerReasons mismatch: got %v", rec.ChallengerReasons)
	}
	if rec.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}
