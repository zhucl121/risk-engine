// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package orchestrator

import (
	"strings"

	"github.com/zhucl121/risk-engine/internal/engine"
	"github.com/zhucl121/risk-engine/internal/feature"
)

// ParamMapping defines how a downstream input parameter is populated.
// It supports two source syntaxes:
//
//   - "extra.<key>"     — read from DecisionRequest.Extra
//   - "feature.<key>"  — read from the current feature.Map (includes extra.* keys)
//   - bare string       — treated as a literal constant value
//
// Example YAML Step configuration:
//
//	params:
//	  merchant_id: "extra.merchant_id"
//	  product_type: "feature.extra.product_type"
//	  channel: "WEB"                         # literal constant
type ParamMapping map[string]string // downstreamKey → sourceExpression

// ResolvedParams is the concrete key→value map that is passed to rule/model/list calls.
// All values are strings; downstream callers type-assert via feature.Map helpers.
type ResolvedParams map[string]feature.Value

// resolve evaluates a ParamMapping against the request and feature.Map,
// returning a ResolvedParams that downstream dispatch functions can consume.
func (pm ParamMapping) resolve(req *engine.DecisionRequest, features feature.Map) ResolvedParams {
	if len(pm) == 0 {
		return nil
	}
	out := make(ResolvedParams, len(pm))
	for destKey, src := range pm {
		out[destKey] = resolveSource(src, req, features)
	}
	return out
}

// resolveSource evaluates one source expression.
func resolveSource(src string, req *engine.DecisionRequest, features feature.Map) feature.Value {
	switch {
	case strings.HasPrefix(src, "extra."):
		k := strings.TrimPrefix(src, "extra.")
		// Look up in the already-injected feature map first (typed value).
		if v, ok := features["extra."+k]; ok {
			return v
		}
		// Fall back to raw Extra string.
		if raw, ok := req.Extra[k]; ok {
			return feature.Value{Kind: feature.KindString, StrVal: raw}
		}
	case strings.HasPrefix(src, "feature."):
		k := strings.TrimPrefix(src, "feature.")
		if v, ok := features[k]; ok {
			return v
		}
	case src == "request.user_id":
		return feature.Value{Kind: feature.KindString, StrVal: req.UserID}
	case src == "request.device_id":
		return feature.Value{Kind: feature.KindString, StrVal: req.DeviceID}
	case src == "request.ip":
		return feature.Value{Kind: feature.KindString, StrVal: req.IP}
	case src == "request.session_id":
		return feature.Value{Kind: feature.KindString, StrVal: req.SessionID}
	}
	// Literal constant.
	return feature.Value{Kind: feature.KindString, StrVal: src}
}

// mergeParams overlays resolved params onto a copy of the feature map,
// returning a new map that downstream callers receive.
// Resolved params take precedence over existing feature values for the same key.
func mergeParams(base feature.Map, params ResolvedParams) feature.Map {
	if len(params) == 0 {
		return base
	}
	merged := make(feature.Map, len(base)+len(params))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range params {
		merged[k] = v
	}
	return merged
}
