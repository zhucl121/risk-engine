// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"

	"github.com/yourorg/riskengine/pkg/dsl"
)

// velocityFn implements velocity(event string, window string) → int.
// The key is derived from the current request's UserID.
// window syntax: "1m", "5m", "1h", "24h"  (same as pkg/sliding).
func velocityFn(ctx context.Context, rt *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
	if len(args) != 2 {
		return dsl.NilValue(), fmt.Errorf("velocity: expected 2 args, got %d", len(args))
	}
	if args[0].Kind() != dsl.KindString {
		return dsl.NilValue(), fmt.Errorf("velocity: arg 1 (event) must be string")
	}
	if args[1].Kind() != dsl.KindString {
		return dsl.NilValue(), fmt.Errorf("velocity: arg 2 (window) must be string")
	}
	if rt.VelocityFunc == nil {
		return dsl.IntValue(0), nil // graceful degradation
	}
	// Composite key: event + userID so each user has an independent counter.
	key := args[0].Str() + ":" + rt.Request.UserID
	count, err := rt.VelocityFunc.Count(ctx, args[0].Str(), args[1].Str(), key)
	if err != nil {
		return dsl.IntValue(0), fmt.Errorf("velocity(%q, %q): %w", args[0].Str(), args[1].Str(), err)
	}
	return dsl.IntValue(count), nil
}

func registerVelocity(reg *dsl.FunctionRegistry) error {
	return reg.Register(dsl.FuncDef{
		Name:       "velocity",
		Args:       []dsl.ArgKind{dsl.ArgKindString, dsl.ArgKindString},
		ReturnKind: dsl.KindInt,
		Impl:       velocityFn,
	})
}
