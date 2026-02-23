// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package adminv1 provides the management API for risk rules.
// It exposes CRUD operations and DSL validation endpoints under /admin/v1/rules.
package adminv1

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// CreateRuleRequest is the body for POST /admin/v1/rules.
type CreateRuleRequest struct {
	RuleKey        string `json:"rule_key" binding:"required,max=128"`
	Name           string `json:"name"     binding:"required,max=256"`
	GroupName      string `json:"group_name" binding:"required,max=128"`
	SceneCode      string `json:"scene_code"  binding:"required,max=128"`
	Priority       int    `json:"priority"`
	ConditionDSL   string `json:"condition_dsl"  binding:"required"`
	ConditionAST   any    `json:"condition_ast"`   // raw JSON, optional
	ActionDecision string `json:"action_decision" binding:"required,oneof=REJECT MANUAL_REVIEW PASS"`
	ActionRiskCode string `json:"action_risk_code"`
	ActionScore    int    `json:"action_score"`
}

// UpdateRuleRequest is the body for PUT /admin/v1/rules/:id.
// Version is required for optimistic locking.
type UpdateRuleRequest struct {
	Name           string `json:"name"     binding:"required,max=256"`
	GroupName      string `json:"group_name" binding:"required,max=128"`
	SceneCode      string `json:"scene_code"  binding:"required,max=128"`
	Priority       int    `json:"priority"`
	ConditionDSL   string `json:"condition_dsl"  binding:"required"`
	ConditionAST   any    `json:"condition_ast"`
	ActionDecision string `json:"action_decision" binding:"required,oneof=REJECT MANUAL_REVIEW PASS"`
	ActionRiskCode string `json:"action_risk_code"`
	ActionScore    int    `json:"action_score"`
	Version        int    `json:"version" binding:"required,min=1"`
}

// ValidateRequest is the body for POST /admin/v1/rules/validate.
type ValidateRequest struct {
	ConditionDSL string `json:"condition_dsl" binding:"required"`
}

// ListQuery holds query parameters for GET /admin/v1/rules.
type ListQuery struct {
	SceneCode string `form:"scene_code"`
	GroupName string `form:"group_name"`
	Status    *int8  `form:"status"` // nil = all, 0 = disabled, 1 = enabled
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=20"`
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

// RuleResponse is the API representation of a single rule.
type RuleResponse struct {
	ID             int64  `json:"id"`
	RuleKey        string `json:"rule_key"`
	Name           string `json:"name"`
	GroupName      string `json:"group_name"`
	SceneCode      string `json:"scene_code"`
	Priority       int    `json:"priority"`
	ConditionDSL   string `json:"condition_dsl"`
	ConditionAST   any    `json:"condition_ast,omitempty"`
	ActionDecision string `json:"action_decision"`
	ActionRiskCode string `json:"action_risk_code,omitempty"`
	ActionScore    int    `json:"action_score"`
	Status         int8   `json:"status"`
	Version        int    `json:"version"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// ValidateResponse is returned by POST /admin/v1/rules/validate.
type ValidateResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

// ReloadResponse is returned by POST /admin/v1/rules/reload.
type ReloadResponse struct {
	Message string `json:"message"`
}

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
