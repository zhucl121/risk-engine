// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package orchestrator

import (
	"strconv"
	"strings"

	"github.com/yourorg/riskengine/internal/engine"
	"github.com/yourorg/riskengine/internal/feature"
)

// ExtraFieldType declares the intended type for one Extra key.
// When an ExtraSchema is attached to a PolicySet or Step, the engine
// coerces string values from DecisionRequest.Extra before injecting them
// into the feature.Map.
//
// Without a schema entry the value is always injected as KindString.
type ExtraFieldType string

const (
	ExtraTypeString ExtraFieldType = "string" // default when no schema entry
	ExtraTypeInt    ExtraFieldType = "int"
	ExtraTypeFloat  ExtraFieldType = "float"
	ExtraTypeBool   ExtraFieldType = "bool"
)

// ExtraSchema maps Extra key names to their declared types.
// Keys that are absent from the schema are treated as string.
type ExtraSchema map[string]ExtraFieldType

// injectExtra merges DecisionRequest.Extra into the feature.Map under the
// "extra." namespace.  Values are type-coerced according to schema; unknown
// keys default to string.
//
// Example:
//
//	Extra = {"amount_usd": "9.99", "is_vip": "true"}
//	schema = {"amount_usd": "float", "is_vip": "bool"}
//	→ features["extra.amount_usd"] = Value{Kind: KindFloat, FltVal: 9.99}
//	→ features["extra.is_vip"]     = Value{Kind: KindBool,  BoolVal: true}
func injectExtra(req *engine.DecisionRequest, schema ExtraSchema, dest feature.Map) {
	for k, raw := range req.Extra {
		fkey := "extra." + k
		typ := schema[k] // zero-value "" → treated as string below
		dest[fkey] = coerceExtra(raw, typ)
	}
}

// coerceExtra converts a raw string value to a typed feature.Value.
// On parse failure it falls back to KindString so the request is never rejected.
func coerceExtra(raw string, typ ExtraFieldType) feature.Value {
	switch typ {
	case ExtraTypeInt:
		if n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64); err == nil {
			return feature.Value{Kind: feature.KindInt, IntVal: n}
		}
	case ExtraTypeFloat:
		if f, err := strconv.ParseFloat(strings.TrimSpace(raw), 64); err == nil {
			return feature.Value{Kind: feature.KindFloat, FltVal: f}
		}
	case ExtraTypeBool:
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "true", "1", "yes":
			return feature.Value{Kind: feature.KindBool, BoolVal: true}
		case "false", "0", "no":
			return feature.Value{Kind: feature.KindBool, BoolVal: false}
		}
	}
	// Default: treat as string (also the fallback on parse error).
	return feature.Value{Kind: feature.KindString, StrVal: raw}
}
