// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"
	"math"

	"github.com/zhucl121/risk-engine/pkg/dsl"
)

func registerMath(reg *dsl.FunctionRegistry) error {
	numeric1 := func(name string, fn func(float64) float64) func(*dsl.FunctionRegistry) error {
		return func(r *dsl.FunctionRegistry) error {
			return r.RegisterFunc(name, func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
				if len(args) != 1 {
					return dsl.NilValue(), fmt.Errorf("%s() requires 1 arg", name)
				}
				v, ok := args[0].Numeric()
				if !ok {
					return dsl.NilValue(), fmt.Errorf("%s(): arg must be numeric", name)
				}
				return dsl.FloatValue(fn(v)), nil
			})
		}
	}

	fns := []func(*dsl.FunctionRegistry) error{
		numeric1("abs", math.Abs),
		numeric1("ceil", math.Ceil),
		numeric1("floor", math.Floor),
		numeric1("round", math.Round),
		numeric1("sqrt", math.Sqrt),

		// min(a, b) → numeric
		func(r *dsl.FunctionRegistry) error {
			return r.RegisterFunc("min", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
				if len(args) != 2 {
					return dsl.NilValue(), fmt.Errorf("min() requires 2 args")
				}
				a, aok := args[0].Numeric()
				b, bok := args[1].Numeric()
				if !aok || !bok {
					return dsl.NilValue(), fmt.Errorf("min(): args must be numeric")
				}
				if a < b {
					return dsl.FloatValue(a), nil
				}
				return dsl.FloatValue(b), nil
			})
		},
		// max(a, b) → numeric
		func(r *dsl.FunctionRegistry) error {
			return r.RegisterFunc("max", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
				if len(args) != 2 {
					return dsl.NilValue(), fmt.Errorf("max() requires 2 args")
				}
				a, aok := args[0].Numeric()
				b, bok := args[1].Numeric()
				if !aok || !bok {
					return dsl.NilValue(), fmt.Errorf("max(): args must be numeric")
				}
				if a > b {
					return dsl.FloatValue(a), nil
				}
				return dsl.FloatValue(b), nil
			})
		},
		// clamp(v, lo, hi) → numeric
		func(r *dsl.FunctionRegistry) error {
			return r.RegisterFunc("clamp", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
				if len(args) != 3 {
					return dsl.NilValue(), fmt.Errorf("clamp() requires 3 args")
				}
				v, vok := args[0].Numeric()
				lo, lok := args[1].Numeric()
				hi, hok := args[2].Numeric()
				if !vok || !lok || !hok {
					return dsl.NilValue(), fmt.Errorf("clamp(): args must be numeric")
				}
				if v < lo {
					return dsl.FloatValue(lo), nil
				}
				if v > hi {
					return dsl.FloatValue(hi), nil
				}
				return dsl.FloatValue(v), nil
			})
		},
	}

	for _, fn := range fns {
		if err := fn(reg); err != nil {
			return fmt.Errorf("builtins.math: %w", err)
		}
	}
	return nil
}
