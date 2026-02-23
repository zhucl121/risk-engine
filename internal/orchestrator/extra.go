// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package orchestrator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yourorg/riskengine/internal/engine"
	"github.com/yourorg/riskengine/internal/feature"
	"github.com/yourorg/riskengine/internal/scene"
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

// specsToSchema converts a slice of ExtraParamSpec (loaded from the DB)
// into an ExtraSchema map for type-coercion in injectExtra.
// Only specs with status=1 (active) are included.
func specsToSchema(specs []scene.ExtraParamSpec) ExtraSchema {
	schema := make(ExtraSchema, len(specs))
	for _, s := range specs {
		if s.Status != 1 {
			continue
		}
		switch s.ParamType {
		case "int":
			schema[s.ParamKey] = ExtraTypeInt
		case "float":
			schema[s.ParamKey] = ExtraTypeFloat
		case "bool":
			schema[s.ParamKey] = ExtraTypeBool
		default:
			schema[s.ParamKey] = ExtraTypeString
		}
	}
	return schema
}

// mergeSchemas merges two schemas, with override taking precedence over base.
// Used to overlay DB specs on top of the static YAML ExtraSchema.
func mergeSchemas(base, override ExtraSchema) ExtraSchema {
	merged := make(ExtraSchema, len(base)+len(override))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range override {
		merged[k] = v
	}
	return merged
}

// ErrMissingRequiredExtra is returned when a required Extra field is absent.
type ErrMissingRequiredExtra struct {
	SceneCode string
	ParamKey  string
}

func (e *ErrMissingRequiredExtra) Error() string {
	return fmt.Sprintf("extra field %q is required for scene %q", e.ParamKey, e.SceneCode)
}

// validateAndFillExtra validates required fields and fills default values for
// optional fields that are absent in req.Extra.
//
//   - If a required spec's key is missing from req.Extra, ErrMissingRequiredExtra
//     is returned and the request must be rejected.
//   - If an optional spec's key is absent and a default is configured, the
//     default value is written into req.Extra so that injectExtra sees it.
//
// The function mutates req.Extra in-place for default-value filling.
func validateAndFillExtra(req *engine.DecisionRequest, specs []scene.ExtraParamSpec, sceneCode string) error {
	for _, spec := range specs {
		if spec.Status != 1 {
			continue
		}
		_, present := req.Extra[spec.ParamKey]
		if !present {
			if spec.Required {
				return &ErrMissingRequiredExtra{SceneCode: sceneCode, ParamKey: spec.ParamKey}
			}
			if spec.HasDefault() {
				if req.Extra == nil {
					req.Extra = make(map[string]string)
				}
				req.Extra[spec.ParamKey] = spec.DefaultVal
			}
		}
	}
	return nil
}

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
