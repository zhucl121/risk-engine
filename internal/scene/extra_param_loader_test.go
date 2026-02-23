// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package scene_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhucl121/risk-engine/internal/scene"
)

// stubRepo is an in-memory ExtraParamRepository for testing.
type stubRepo struct {
	byScene map[string][]*scene.ExtraParamSpec
	// callCount tracks how many times ListByScene was called.
	callCount int
}

func newStubRepo(specs ...*scene.ExtraParamSpec) *stubRepo {
	r := &stubRepo{byScene: make(map[string][]*scene.ExtraParamSpec)}
	for _, s := range specs {
		r.byScene[s.SceneCode] = append(r.byScene[s.SceneCode], s)
	}
	return r
}

func (r *stubRepo) ListByScene(_ context.Context, code string) ([]*scene.ExtraParamSpec, error) {
	r.callCount++
	return r.byScene[code], nil
}

func (r *stubRepo) ListUpdatedSince(_ context.Context, _ time.Time) ([]*scene.ExtraParamSpec, error) {
	return nil, nil
}

func (r *stubRepo) GetByKey(_ context.Context, code, key string) (*scene.ExtraParamSpec, error) {
	for _, s := range r.byScene[code] {
		if s.ParamKey == key {
			return s, nil
		}
	}
	return nil, errors.New("not found")
}

func (r *stubRepo) Upsert(_ context.Context, spec *scene.ExtraParamSpec) error {
	r.byScene[spec.SceneCode] = append(r.byScene[spec.SceneCode], spec)
	return nil
}

func (r *stubRepo) Delete(_ context.Context, code, key string) error {
	return nil
}

// TestLoader_LoadFirstCall verifies that the first call performs a DB fetch.
func TestLoader_LoadFirstCall(t *testing.T) {
	spec := &scene.ExtraParamSpec{
		SceneCode:  "pay",
		ParamKey:   "merchant_id",
		ParamType:  "string",
		Required:   true,
		DefaultVal: "",
		Status:     1,
	}
	repo := newStubRepo(spec)
	loader := scene.NewExtraParamLoader(repo, 0, nil)

	specs, err := loader.Load(context.Background(), "pay")
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "merchant_id", specs[0].ParamKey)
	assert.True(t, specs[0].Required)
	assert.Equal(t, 1, repo.callCount, "DB should be queried once on first call")
}

// TestLoader_LoadCachesResult verifies that repeated Load calls do not
// re-query the DB.
func TestLoader_LoadCachesResult(t *testing.T) {
	repo := newStubRepo(&scene.ExtraParamSpec{
		SceneCode: "pay", ParamKey: "k", ParamType: "string", Status: 1,
	})
	loader := scene.NewExtraParamLoader(repo, 0, nil)

	_, _ = loader.Load(context.Background(), "pay")
	_, _ = loader.Load(context.Background(), "pay")
	_, _ = loader.Load(context.Background(), "pay")

	assert.Equal(t, 1, repo.callCount, "DB should be queried only once; subsequent calls served from cache")
}

// TestLoader_InvalidateForcesRefetch verifies that Invalidate clears the cache,
// causing the next Load to query the DB again.
func TestLoader_InvalidateForcesRefetch(t *testing.T) {
	repo := newStubRepo(&scene.ExtraParamSpec{
		SceneCode: "pay", ParamKey: "k", ParamType: "string", Status: 1,
	})
	loader := scene.NewExtraParamLoader(repo, 0, nil)

	_, _ = loader.Load(context.Background(), "pay")
	assert.Equal(t, 1, repo.callCount)

	loader.Invalidate("pay")

	_, _ = loader.Load(context.Background(), "pay")
	assert.Equal(t, 2, repo.callCount, "DB should be re-queried after cache invalidation")
}

// TestLoader_EmptyScene verifies that Load returns an empty slice (not an error)
// for scenes with no configured specs.
func TestLoader_EmptyScene(t *testing.T) {
	loader := scene.NewExtraParamLoader(newStubRepo(), 0, nil)

	specs, err := loader.Load(context.Background(), "unknown_scene")
	require.NoError(t, err)
	assert.Empty(t, specs, "should return empty slice for unknown scene")
}

// TestLoader_MultipleScenesCached verifies that each scene has its own
// independent cache entry.
func TestLoader_MultipleScenesCached(t *testing.T) {
	repo := newStubRepo(
		&scene.ExtraParamSpec{SceneCode: "a", ParamKey: "k1", Status: 1},
		&scene.ExtraParamSpec{SceneCode: "b", ParamKey: "k2", Status: 1},
	)
	loader := scene.NewExtraParamLoader(repo, 0, nil)

	specsA, _ := loader.Load(context.Background(), "a")
	specsB, _ := loader.Load(context.Background(), "b")

	require.Len(t, specsA, 1)
	require.Len(t, specsB, 1)
	assert.Equal(t, "k1", specsA[0].ParamKey)
	assert.Equal(t, "k2", specsB[0].ParamKey)
	assert.Equal(t, 2, repo.callCount, "each scene should be fetched once")
}

// TestExtraParamSpec_HasDefault verifies the HasDefault helper.
func TestExtraParamSpec_HasDefault(t *testing.T) {
	assert.True(t, (&scene.ExtraParamSpec{DefaultVal: "GOODS"}).HasDefault())
	assert.False(t, (&scene.ExtraParamSpec{DefaultVal: ""}).HasDefault())
}
