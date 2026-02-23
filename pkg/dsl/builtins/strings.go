// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/zhucl121/risk-engine/pkg/dsl"
)

// regexpCache caches compiled regular expressions for the match() function.
var (
	regexpCache   = sync.Map{}
)

func compileRegexp(pattern string) (*regexp.Regexp, error) {
	if v, ok := regexpCache.Load(pattern); ok {
		return v.(*regexp.Regexp), nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	regexpCache.Store(pattern, re)
	return re, nil
}

func registerStrings(reg *dsl.FunctionRegistry) error {
	fns := []struct {
		name string
		impl dsl.FuncImpl
	}{
		// contains(s, substr) → bool
		{"contains", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 2 {
				return dsl.NilValue(), fmt.Errorf("contains() requires 2 args")
			}
			s, p := args[0].String(), args[1].String()
			return dsl.BoolValue(strings.Contains(s, p)), nil
		}},
		// startsWith(s, prefix) → bool
		{"startsWith", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 2 {
				return dsl.NilValue(), fmt.Errorf("startsWith() requires 2 args")
			}
			return dsl.BoolValue(strings.HasPrefix(args[0].String(), args[1].String())), nil
		}},
		// endsWith(s, suffix) → bool
		{"endsWith", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 2 {
				return dsl.NilValue(), fmt.Errorf("endsWith() requires 2 args")
			}
			return dsl.BoolValue(strings.HasSuffix(args[0].String(), args[1].String())), nil
		}},
		// match(s, pattern) → bool  (RE2 regexp)
		{"match", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 2 {
				return dsl.NilValue(), fmt.Errorf("match() requires 2 args")
			}
			re, err := compileRegexp(args[1].String())
			if err != nil {
				return dsl.NilValue(), fmt.Errorf("match(): invalid regexp %q: %w", args[1].String(), err)
			}
			return dsl.BoolValue(re.MatchString(args[0].String())), nil
		}},
		// lower(s) → string
		{"lower", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 1 {
				return dsl.NilValue(), fmt.Errorf("lower() requires 1 arg")
			}
			return dsl.StringValue(strings.ToLower(args[0].String())), nil
		}},
		// upper(s) → string
		{"upper", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 1 {
				return dsl.NilValue(), fmt.Errorf("upper() requires 1 arg")
			}
			return dsl.StringValue(strings.ToUpper(args[0].String())), nil
		}},
		// trim(s) → string
		{"trim", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 1 {
				return dsl.NilValue(), fmt.Errorf("trim() requires 1 arg")
			}
			return dsl.StringValue(strings.TrimSpace(args[0].String())), nil
		}},
		// strlen(s) → int
		{"strlen", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 1 {
				return dsl.NilValue(), fmt.Errorf("strlen() requires 1 arg")
			}
			return dsl.IntValue(int64(len([]rune(args[0].String())))), nil
		}},
		// isEmpty(s) → bool
		{"isEmpty", func(_ context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
			if len(args) != 1 {
				return dsl.NilValue(), fmt.Errorf("isEmpty() requires 1 arg")
			}
			return dsl.BoolValue(args[0].String() == ""), nil
		}},
	}

	for _, f := range fns {
		if err := reg.RegisterFunc(f.name, f.impl); err != nil {
			return fmt.Errorf("builtins.strings: %w", err)
		}
	}
	return nil
}
