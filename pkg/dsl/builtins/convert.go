// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/zhucl121/risk-engine/pkg/dsl"
)

func registerConvert(reg *dsl.FunctionRegistry) error {
	fns := []struct {
		name string
		impl dsl.FuncImpl
	}{
		// toInt(v) → int  (converts string or float to int64; 0 on failure)
		{"toInt", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 1 {
				return dsl.NilValue(), fmt.Errorf("toInt() requires 1 arg")
			}
			switch args[0].Kind() {
			case dsl.KindInt:
				return args[0], nil
			case dsl.KindFloat:
				return dsl.IntValue(int64(args[0].Float())), nil
			case dsl.KindBool:
				if args[0].Bool() {
					return dsl.IntValue(1), nil
				}
				return dsl.IntValue(0), nil
			case dsl.KindString:
				n, err := strconv.ParseInt(strings.TrimSpace(args[0].String()), 10, 64)
				if err != nil {
					f, ferr := strconv.ParseFloat(strings.TrimSpace(args[0].String()), 64)
					if ferr == nil {
						return dsl.IntValue(int64(f)), nil
					}
					return dsl.IntValue(0), nil
				}
				return dsl.IntValue(n), nil
			}
			return dsl.IntValue(0), nil
		}},
		// toFloat(v) → float
		{"toFloat", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 1 {
				return dsl.NilValue(), fmt.Errorf("toFloat() requires 1 arg")
			}
			switch args[0].Kind() {
			case dsl.KindFloat:
				return args[0], nil
			case dsl.KindInt:
				return dsl.FloatValue(float64(args[0].Int())), nil
			case dsl.KindString:
				f, err := strconv.ParseFloat(strings.TrimSpace(args[0].String()), 64)
				if err != nil {
					return dsl.FloatValue(0), nil
				}
				return dsl.FloatValue(f), nil
			}
			return dsl.FloatValue(0), nil
		}},
		// toString(v) → string
		{"toString", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 1 {
				return dsl.NilValue(), fmt.Errorf("toString() requires 1 arg")
			}
			return dsl.StringValue(args[0].String()), nil
		}},
		// toBool(v) → bool  ("true"/"1"/"yes" → true)
		{"toBool", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 1 {
				return dsl.NilValue(), fmt.Errorf("toBool() requires 1 arg")
			}
			switch args[0].Kind() {
			case dsl.KindBool:
				return args[0], nil
			case dsl.KindInt:
				return dsl.BoolValue(args[0].Int() != 0), nil
			case dsl.KindString:
				switch strings.ToLower(strings.TrimSpace(args[0].String())) {
				case "true", "1", "yes":
					return dsl.BoolValue(true), nil
				}
				return dsl.BoolValue(false), nil
			}
			return dsl.BoolValue(false), nil
		}},
		// isNull(v) → bool
		{"isNull", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 1 {
				return dsl.NilValue(), fmt.Errorf("isNull() requires 1 arg")
			}
			return dsl.BoolValue(args[0].Kind() == dsl.KindNil), nil
		}},
		// coalesce(a, b) → first non-nil value
		{"coalesce", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			for _, a := range args {
				if a.Kind() != dsl.KindNil {
					return a, nil
				}
			}
			return dsl.NilValue(), nil
		}},
		// ifThen(cond, trueVal, falseVal) → value  (functional ternary)
		{"ifThen", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 3 {
				return dsl.NilValue(), fmt.Errorf("ifThen() requires 3 args")
			}
			if args[0].Kind() != dsl.KindBool {
				return dsl.NilValue(), fmt.Errorf("ifThen(): first arg must be bool")
			}
			if args[0].Bool() {
				return args[1], nil
			}
			return args[2], nil
		}},
	}

	for _, f := range fns {
		if err := reg.RegisterFunc(f.name, f.impl); err != nil {
			return fmt.Errorf("builtins.convert: %w", err)
		}
	}
	return nil
}
