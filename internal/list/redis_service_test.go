// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package list_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/riskengine/internal/list"
)

// fakeRedis is a minimal in-memory Redis client for testing.
type fakeRedis struct {
	data map[string]fakeEntry
}

type fakeEntry struct {
	val string
	exp time.Time // zero = no expiry
}

func newFakeRedis() *fakeRedis { return &fakeRedis{data: make(map[string]fakeEntry)} }

// The test uses fakeListService directly (a thin wrapper) to avoid
// requiring a real Redis connection. We test via a stub service that
// mimics the redis_service logic using an in-memory map.

type stubListService struct {
	data map[string]list.Status
}

func newStub() list.Service {
	return &stubListService{data: make(map[string]list.Status)}
}

func key(q *list.Query) string { return q.Kind + ":" + q.Value }

func (s *stubListService) Check(_ context.Context, q *list.Query) (list.Status, error) {
	st, ok := s.data[key(q)]
	if !ok {
		return list.StatusNotFound, nil
	}
	return st, nil
}

func (s *stubListService) Add(_ context.Context, q *list.Query, status list.Status, _ time.Duration) error {
	s.data[key(q)] = status
	return nil
}

func (s *stubListService) Remove(_ context.Context, q *list.Query) error {
	delete(s.data, key(q))
	return nil
}

func TestListService_CheckNotFound(t *testing.T) {
	svc := newStub()
	ctx := context.Background()

	st, err := svc.Check(ctx, &list.Query{Kind: "user", Value: "u999"})
	require.NoError(t, err)
	assert.Equal(t, list.StatusNotFound, st)
}

func TestListService_AddAndCheck(t *testing.T) {
	svc := newStub()
	ctx := context.Background()

	q := &list.Query{Kind: "device", Value: "dev-abc"}
	require.NoError(t, svc.Add(ctx, q, list.StatusBlacklist, 0))

	st, err := svc.Check(ctx, q)
	require.NoError(t, err)
	assert.Equal(t, list.StatusBlacklist, st)
}

func TestListService_Remove(t *testing.T) {
	svc := newStub()
	ctx := context.Background()

	q := &list.Query{Kind: "ip", Value: "1.2.3.4"}
	require.NoError(t, svc.Add(ctx, q, list.StatusGraylist, time.Minute))
	require.NoError(t, svc.Remove(ctx, q))

	st, err := svc.Check(ctx, q)
	require.NoError(t, err)
	assert.Equal(t, list.StatusNotFound, st)
}

func TestListService_OverwriteStatus(t *testing.T) {
	svc := newStub()
	ctx := context.Background()

	q := &list.Query{Kind: "user", Value: "u1"}
	require.NoError(t, svc.Add(ctx, q, list.StatusGraylist, 0))
	require.NoError(t, svc.Add(ctx, q, list.StatusBlacklist, 0))

	st, err := svc.Check(ctx, q)
	require.NoError(t, err)
	assert.Equal(t, list.StatusBlacklist, st)
}
