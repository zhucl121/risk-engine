// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package orchestrator

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// CanaryConfig enables deterministic, hash-based traffic splitting for
// gradual policy rollout (grey release / canary deployment).
//
// Unlike ABTestConfig which uses random sampling (stateless, may flip per
// request), CanaryConfig uses a stable hash of the user's identity so the
// same user always lands in the same bucket.  This eliminates the "jitter"
// problem where a user sees different decisions on consecutive requests.
//
// Routing algorithm:
//
//	bucket = murmurLike(salt + ":" + hashKey) % 100
//	if bucket < TrafficPct → use CanaryPipeline
//	else                   → use main Pipeline
//
// YAML example:
//
//	canary:
//	  enabled: true
//	  canaryVersion: "v2.1.0"
//	  trafficPct: 10           # 0-100 integer; 10 = 10 % of users
//	  hashKey: userID          # userID | deviceID | sessionID | ip | extra.<key>
//	  salt: "payment_v2"       # per-experiment salt to avoid bucket correlation
//	  canaryPipeline:
//	    - name: model_v2
//	      kind: MODEL
//	      models: [fraud_v2]
type CanaryConfig struct {
	Enabled bool
	// CanaryVersion is a human-readable tag for audit traceability.
	CanaryVersion string
	// TrafficPct is the percentage of traffic (0–100) routed to the canary pipeline.
	// 0 = disabled, 100 = all traffic.
	TrafficPct int
	// HashKey determines which request field is used for bucket assignment.
	// Supported values: "userID", "deviceID", "sessionID", "ip", "extra.<key>"
	// Default: "userID"
	HashKey string
	// Salt is mixed into the hash to allow independent experiments without bucket
	// correlation.  Different experiments MUST use different salts.
	Salt string
	// CanaryPipeline is the alternative step sequence for canary traffic.
	// When empty, the main Pipeline is reused (useful for config-only canaries).
	CanaryPipeline []Step
}

// canaryBucket computes a deterministic bucket number [0, 100) for the given key.
// Uses SHA-256 for uniform distribution and collision resistance.
func canaryBucket(salt, key string) int {
	input := salt + ":" + key
	sum := sha256.Sum256([]byte(input))
	// Take first 8 bytes as uint64, then mod 100.
	v := binary.BigEndian.Uint64(sum[:8])
	return int(v % 100)
}

// resolveHashKey extracts the routing key from the request based on cfg.HashKey.
func resolveHashKey(cfg CanaryConfig, req interface{ GetField(string) string }) string {
	switch cfg.HashKey {
	case "", "userID":
		return req.GetField("userID")
	case "deviceID":
		return req.GetField("deviceID")
	case "sessionID":
		return req.GetField("sessionID")
	case "ip":
		return req.GetField("ip")
	default:
		// "extra.<key>" form.
		return req.GetField(cfg.HashKey)
	}
}

// canaryHashKeyLabel returns a display-friendly label for the hash key.
func canaryHashKeyLabel(hashKey string) string {
	if hashKey == "" {
		return "userID"
	}
	return hashKey
}

// canaryRoutingKey extracts the actual string value used for hash routing
// from a DecisionRequest.  Avoids importing engine in this file by using
// field-name strings passed from the caller.
func canaryRoutingKey(hashKey, userID, deviceID, sessionID, ip string, extra map[string]string) string {
	switch hashKey {
	case "", "userID":
		return userID
	case "deviceID":
		return deviceID
	case "sessionID":
		return sessionID
	case "ip":
		return ip
	default:
		// "extra.<key>" form.
		k := hashKey
		if len(k) > 6 && k[:6] == "extra." {
			k = k[6:]
		}
		return extra[k]
	}
}

// inCanary returns true when the request should be routed to the canary pipeline.
func inCanary(cfg *CanaryConfig, userID, deviceID, sessionID, ip string, extra map[string]string) bool {
	if cfg == nil || !cfg.Enabled || cfg.TrafficPct <= 0 {
		return false
	}
	if cfg.TrafficPct >= 100 {
		return true
	}
	key := canaryRoutingKey(cfg.HashKey, userID, deviceID, sessionID, ip, extra)
	if key == "" {
		key = fmt.Sprintf("fallback-%s-%s", userID, deviceID) // best-effort when key is empty
	}
	bucket := canaryBucket(cfg.Salt, key)
	return bucket < cfg.TrafficPct
}
