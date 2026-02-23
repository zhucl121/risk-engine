// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"

	"github.com/zhucl121/risk-engine/pkg/dsl"
)

// withinFn implements within(v, lo, hi) → bool: returns lo <= v <= hi.
// All three arguments must be numeric (int or float).
func withinFn(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
	if len(args) != 3 {
		return dsl.NilValue(), fmt.Errorf("within: expected 3 args, got %d", len(args))
	}
	v, ok1 := args[0].Numeric()
	lo, ok2 := args[1].Numeric()
	hi, ok3 := args[2].Numeric()
	if !ok1 || !ok2 || !ok3 {
		return dsl.NilValue(), fmt.Errorf("within: all arguments must be numeric")
	}
	return dsl.BoolValue(v >= lo && v <= hi), nil
}

func registerWithin(reg *dsl.FunctionRegistry) error {
	return reg.Register(dsl.FuncDef{
		Name:       "within",
		Args:       []dsl.ArgKind{dsl.ArgKindNumber, dsl.ArgKindNumber, dsl.ArgKindNumber},
		ReturnKind: dsl.KindBool,
		Impl:       withinFn,
	})
}
