// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"

	"github.com/yourorg/riskengine/pkg/dsl"
)

// inListFn implements inList(listName string, value string) → bool.
// It delegates to the Runtime.ListChecker injected at startup.
func inListFn(ctx context.Context, rt *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
	if len(args) != 2 {
		return dsl.NilValue(), fmt.Errorf("inList: expected 2 args, got %d", len(args))
	}
	if args[0].Kind() != dsl.KindString {
		return dsl.NilValue(), fmt.Errorf("inList: arg 1 (listName) must be string")
	}
	if args[1].Kind() != dsl.KindString {
		return dsl.NilValue(), fmt.Errorf("inList: arg 2 (value) must be string")
	}
	if rt.ListChecker == nil {
		return dsl.BoolValue(false), nil // graceful degradation when list service unavailable
	}
	hit, err := rt.ListChecker.InList(ctx, args[0].Str(), args[1].Str())
	if err != nil {
		// On list service error, return false (fail-open) and surface as non-fatal.
		return dsl.BoolValue(false), fmt.Errorf("inList(%q, %q): %w", args[0].Str(), args[1].Str(), err)
	}
	return dsl.BoolValue(hit), nil
}

func registerInList(reg *dsl.FunctionRegistry) error {
	return reg.Register(dsl.FuncDef{
		Name:       "inList",
		Args:       []dsl.ArgKind{dsl.ArgKindString, dsl.ArgKindString},
		ReturnKind: dsl.KindBool,
		Impl:       inListFn,
	})
}
