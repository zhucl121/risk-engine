// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package expr wraps antonmedv/expr to provide a compile-once, execute-many
// expression evaluator for rule conditions. Expressions are compiled at rule
// load time and the resulting Program is reused across evaluations, giving
// sub-microsecond evaluation latency for simple conditions.
package expr

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

// Program is a compiled, reusable expression ready for fast evaluation.
type Program struct {
	p    *vm.Program
	src  string
}

// Compile parses and type-checks the expression against env.
// env should be a struct or map that reflects the evaluation environment;
// it is used only for type checking and is not retained after Compile returns.
func Compile(expression string, env any) (*Program, error) {
	p, err := expr.Compile(expression, expr.Env(env))
	if err != nil {
		return nil, fmt.Errorf("expr compile %q: %w", expression, err)
	}
	return &Program{p: p, src: expression}, nil
}

// Run evaluates the compiled expression against env and returns the result.
// env must be compatible with the environment used at Compile time.
func (p *Program) Run(env any) (any, error) {
	result, err := expr.Run(p.p, env)
	if err != nil {
		return nil, fmt.Errorf("expr run %q: %w", p.src, err)
	}
	return result, nil
}

// RunBool is a convenience wrapper that asserts the result is a bool.
func (p *Program) RunBool(env any) (bool, error) {
	result, err := p.Run(env)
	if err != nil {
		return false, err
	}
	b, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("expr %q: expected bool, got %T", p.src, result)
	}
	return b, nil
}

// Source returns the original expression string.
func (p *Program) Source() string { return p.src }
