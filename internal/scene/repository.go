// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package scene

import (
	"context"
	"time"
)

// ExtraParamRepository is the data access interface for scene_extra_params.
// All implementations must be safe for concurrent use.
type ExtraParamRepository interface {
	// ListByScene returns all active (status=1) specs for the given scene.
	ListByScene(ctx context.Context, sceneCode string) ([]*ExtraParamSpec, error)

	// ListUpdatedSince returns specs across all scenes that were updated after
	// the given time, used by the background hot-reload watcher.
	ListUpdatedSince(ctx context.Context, since time.Time) ([]*ExtraParamSpec, error)

	// GetByKey returns the spec for (sceneCode, paramKey), or an error if not found.
	GetByKey(ctx context.Context, sceneCode, paramKey string) (*ExtraParamSpec, error)

	// Upsert creates or updates a spec.
	// On conflict (scene_code, param_key) the existing row is updated.
	Upsert(ctx context.Context, spec *ExtraParamSpec) error

	// Delete soft-deletes the spec (sets status=0).
	Delete(ctx context.Context, sceneCode, paramKey string) error
}
