// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package dsl

import (
	"context"
	"fmt"
)

// evalFn is the type of a compiled node: a Go closure that evaluates
// immediately against the given runtime context.
type evalFn func(ctx context.Context, rt *Runtime) (Value, error)

// Program is a compiled, reusable DSL condition.
// It holds a root Go closure produced by the code generator; calling Run
// does not invoke any ANTLR4 code — only pure Go function calls.
//
// Program values are immutable after construction and safe for concurrent use.
type Program struct {
	src  string  // original DSL string, for debugging
	eval evalFn  // root closure produced by codegen
}

// Source returns the original DSL condition string.
func (p *Program) Source() string { return p.src }

// Run evaluates the compiled condition against the given runtime.
// It returns (true/false, nil) on success, or (false, error) on evaluation failure.
//
// rt MUST NOT be nil. Callers should obtain rt via AcquireRuntime / ReleaseRuntime.
func (p *Program) Run(ctx context.Context, rt *Runtime) (bool, error) {
	if ctx.Err() != nil {
		return false, fmt.Errorf("dsl: context cancelled before evaluation: %w", ctx.Err())
	}
	v, err := p.eval(ctx, rt)
	if err != nil {
		return false, fmt.Errorf("dsl: evaluation error in %q: %w", p.src, err)
	}
	if v.Kind() != KindBool {
		return false, fmt.Errorf("dsl: condition %q evaluated to kind %d, expected bool", p.src, v.Kind())
	}
	return v.Bool(), nil
}
