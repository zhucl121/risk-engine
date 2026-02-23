// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package orchestrator

// ChampionChallengerConfig enables simultaneous execution of a production
// (champion) pipeline and one or more experimental (challenger) pipelines.
//
// Design goals:
//  1. The champion's decision is ALWAYS returned to the caller — challengers
//     never influence real business outcomes.
//  2. Both champion and challenger results are written to the audit log under
//     the "champion_challenger_audit" key, enabling statistical comparison
//     without production risk.
//  3. Traffic steering uses the same deterministic hash as CanaryConfig, so
//     the challenger consistently sees the same user population across
//     requests (important for cohort analysis).
//  4. Multiple challengers can run simultaneously; they are always executed
//     concurrently with each other (and with the champion).
//
// Typical workflow:
//
//	Champion (current production policy)   ← always wins, result returned
//	   ↓ (parallel, async)
//	Challenger A (new model / new rules)   ← result written to audit only
//	Challenger B (rule-only baseline)      ← result written to audit only
//
// After enough samples are collected, analysts promote the best challenger
// to champion via a configuration change — no code deployment needed.
//
// YAML example:
//
//	championChallenger:
//	  enabled: true
//	  experimentID: "fraud_model_v3_eval"
//	  challengers:
//	    - challengerID: "model_v3"
//	      trafficPct: 20        # 20 % of users see this challenger executed
//	      hashKey: userID
//	      salt: "cc_fraud_v3"
//	      pipeline:
//	        - name: model_v3
//	          kind: MODEL
//	          models: [fraud_v3]
//	    - challengerID: "rule_only_baseline"
//	      trafficPct: 100       # 100 % = all users, useful as a constant baseline
//	      hashKey: userID
//	      salt: "cc_rule_baseline"
//	      pipeline:
//	        - name: rules
//	          kind: RULE
//	          ruleGroup: baseline_rules
type ChampionChallengerConfig struct {
	Enabled bool
	// ExperimentID groups all challengers under one experiment for reporting.
	ExperimentID string
	// Challengers is the list of challenger variants to run.
	Challengers []ChallengerVariant
}

// ChallengerVariant is one challenger definition within a champion-challenger
// experiment.
type ChallengerVariant struct {
	// ChallengerID is a unique identifier for this variant (used in audit logs).
	ChallengerID string
	// TrafficPct is the percentage of traffic (0–100) for which this challenger
	// is executed.  Uses the same hash-bucket algorithm as CanaryConfig.
	TrafficPct int
	// HashKey is the request field used for deterministic bucket assignment.
	// Supported values: "userID", "deviceID", "sessionID", "ip", "extra.<key>"
	HashKey string
	// Salt is mixed into the hash to prevent bucket correlation with other experiments.
	Salt string
	// Pipeline is the step sequence executed for this challenger.
	Pipeline []Step
}
