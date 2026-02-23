// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package scene manages scene-level configuration that is stored in the
// database and hot-reloaded at runtime, including Extra field parameter
// specifications.
package scene

import "time"

// ExtraParamSpec is the database-level specification for one Extra field
// in a given scene.  It maps 1:1 to a row in scene_extra_params.
type ExtraParamSpec struct {
	ID          int64
	SceneCode   string
	ParamKey    string // Extra field name, e.g. "merchant_id"
	ParamType   string // "string" | "int" | "float" | "bool"
	Required    bool   // true → request is rejected when field is absent
	DefaultVal  string // non-empty → filled in when field is absent and !Required
	Description string
	Status      int8 // 1=active, 0=disabled
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// HasDefault returns true when a non-empty default value is configured.
func (s *ExtraParamSpec) HasDefault() bool {
	return s.DefaultVal != ""
}
