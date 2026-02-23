// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package engine

import "errors"

// Sentinel errors for the engine package.
var (
	// ErrSceneNotFound is returned when no policy is configured for the requested scene.
	ErrSceneNotFound = errors.New("scene not configured")
	// ErrRequestInvalid is returned when the DecisionRequest fails validation.
	ErrRequestInvalid = errors.New("invalid request")
	// ErrTimeout is returned when the decision exceeds the configured deadline.
	ErrTimeout = errors.New("decision timeout")
)
