// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"

	"github.com/zhucl121/risk-engine/pkg/dsl"
)

// modelScoreFn implements modelScore(modelName string) → float.
// It delegates to Runtime.ModelScorer, which routes to model.Registry.
func modelScoreFn(ctx context.Context, rt *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
	if len(args) != 1 {
		return dsl.NilValue(), fmt.Errorf("modelScore: expected 1 arg, got %d", len(args))
	}
	if args[0].Kind() != dsl.KindString {
		return dsl.NilValue(), fmt.Errorf("modelScore: arg 1 (modelName) must be string")
	}
	if rt.ModelScorer == nil {
		return dsl.FloatValue(0), nil // graceful degradation
	}
	score, err := rt.ModelScorer.Score(ctx, args[0].Str(), rt.Features)
	if err != nil {
		return dsl.FloatValue(0), fmt.Errorf("modelScore(%q): %w", args[0].Str(), err)
	}
	return dsl.FloatValue(score), nil
}

func registerModelScore(reg *dsl.FunctionRegistry) error {
	return reg.Register(dsl.FuncDef{
		Name:       "modelScore",
		Args:       []dsl.ArgKind{dsl.ArgKindString},
		ReturnKind: dsl.KindFloat,
		Impl:       modelScoreFn,
	})
}
