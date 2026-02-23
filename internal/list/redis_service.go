// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "riskengine:list:"

// redisService implements Service using Redis as the storage backend.
// Key format: riskengine:list:{kind}:{value} → int value of Status
type redisService struct {
	client redis.UniversalClient
}

// NewRedisService returns a Service backed by the given Redis client.
func NewRedisService(client redis.UniversalClient) Service {
	return &redisService{client: client}
}

func (s *redisService) key(q *Query) string {
	return keyPrefix + q.Kind + ":" + q.Value
}

// Check returns the list status for the entity described by q.
// Returns StatusNotFound when the key does not exist in Redis.
func (s *redisService) Check(ctx context.Context, q *Query) (Status, error) {
	val, err := s.client.Get(ctx, s.key(q)).Result()
	if errors.Is(err, redis.Nil) {
		return StatusNotFound, nil
	}
	if err != nil {
		return StatusNotFound, fmt.Errorf("list.Check %s/%s: %w", q.Kind, q.Value, err)
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return StatusNotFound, fmt.Errorf("list.Check: corrupt value %q for %s/%s", val, q.Kind, q.Value)
	}
	return Status(n), nil
}

// Add inserts or updates an entity's list status.
// A zero ttl means no expiry (the key persists until explicitly removed).
func (s *redisService) Add(ctx context.Context, q *Query, status Status, ttl time.Duration) error {
	val := strconv.Itoa(int(status))
	var err error
	if ttl == 0 {
		err = s.client.Set(ctx, s.key(q), val, 0).Err()
	} else {
		err = s.client.Set(ctx, s.key(q), val, ttl).Err()
	}
	if err != nil {
		return fmt.Errorf("list.Add %s/%s: %w", q.Kind, q.Value, err)
	}
	return nil
}

// Remove deletes an entity from all lists.
func (s *redisService) Remove(ctx context.Context, q *Query) error {
	if err := s.client.Del(ctx, s.key(q)).Err(); err != nil {
		return fmt.Errorf("list.Remove %s/%s: %w", q.Kind, q.Value, err)
	}
	return nil
}
