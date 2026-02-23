// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package dsl

// codegen.go converts an AST into a tree of Go closures (evalFn).
// Each AST node becomes a closure that, when called, evaluates that sub-expression.
// The resulting root closure is stored in Program.eval and called at request time
// with zero ANTLR4 overhead.

import (
	"context"
	"fmt"

	"github.com/yourorg/riskengine/pkg/dsl/ast"
)

// codegenVisitor walks an AST and produces a root evalFn.
type codegenVisitor struct {
	reg *FunctionRegistry
}

// generate is the main entry point: it dispatches on node type.
func (g *codegenVisitor) generate(node ast.Node) (evalFn, error) {
	switch n := node.(type) {
	case *ast.BinaryExpr:
		return g.genBinary(n)
	case *ast.UnaryExpr:
		return g.genUnary(n)
	case *ast.CallExpr:
		return g.genCall(n)
	case *ast.MapIndex:
		return g.genMapIndex(n)
	case *ast.FieldAccess:
		return g.genFieldAccess(n)
	case *ast.Ident:
		return g.genIdent(n)
	case *ast.IntLit:
		v := IntValue(n.Val)
		return func(_ context.Context, _ *Runtime) (Value, error) { return v, nil }, nil
	case *ast.FloatLit:
		v := FloatValue(n.Val)
		return func(_ context.Context, _ *Runtime) (Value, error) { return v, nil }, nil
	case *ast.StringLit:
		v := StringValue(n.Val)
		return func(_ context.Context, _ *Runtime) (Value, error) { return v, nil }, nil
	case *ast.BoolLit:
		v := BoolValue(n.Val)
		return func(_ context.Context, _ *Runtime) (Value, error) { return v, nil }, nil
	default:
		return nil, fmt.Errorf("dsl: unknown AST node type %T", node)
	}
}

func (g *codegenVisitor) genBinary(n *ast.BinaryExpr) (evalFn, error) {
	left, err := g.generate(n.Left)
	if err != nil {
		return nil, err
	}
	right, err := g.generate(n.Right)
	if err != nil {
		return nil, err
	}
	op := n.Op
	line, col := n.Line, n.Col

	return func(ctx context.Context, rt *Runtime) (Value, error) {
		// Short-circuit evaluation for logical operators.
		switch op {
		case "&&":
			lv, err := left(ctx, rt)
			if err != nil {
				return NilValue(), err
			}
			if lv.Kind() != KindBool {
				return NilValue(), &TypeError{Message: "left operand of && is not bool", Line: line, Col: col}
			}
			if !lv.Bool() {
				return BoolValue(false), nil // short-circuit
			}
			rv, err := right(ctx, rt)
			if err != nil {
				return NilValue(), err
			}
			if rv.Kind() != KindBool {
				return NilValue(), &TypeError{Message: "right operand of && is not bool", Line: line, Col: col}
			}
			return BoolValue(rv.Bool()), nil

		case "||":
			lv, err := left(ctx, rt)
			if err != nil {
				return NilValue(), err
			}
			if lv.Kind() != KindBool {
				return NilValue(), &TypeError{Message: "left operand of || is not bool", Line: line, Col: col}
			}
			if lv.Bool() {
				return BoolValue(true), nil // short-circuit
			}
			rv, err := right(ctx, rt)
			if err != nil {
				return NilValue(), err
			}
			if rv.Kind() != KindBool {
				return NilValue(), &TypeError{Message: "right operand of || is not bool", Line: line, Col: col}
			}
			return BoolValue(rv.Bool()), nil
		}

		// Comparison operators.
		lv, err := left(ctx, rt)
		if err != nil {
			return NilValue(), err
		}
		rv, err := right(ctx, rt)
		if err != nil {
			return NilValue(), err
		}
		result, err := compare(op, lv, rv, line, col)
		if err != nil {
			return NilValue(), err
		}
		return BoolValue(result), nil
	}, nil
}

func (g *codegenVisitor) genUnary(n *ast.UnaryExpr) (evalFn, error) {
	operand, err := g.generate(n.Operand)
	if err != nil {
		return nil, err
	}
	line, col := n.Line, n.Col
	return func(ctx context.Context, rt *Runtime) (Value, error) {
		v, err := operand(ctx, rt)
		if err != nil {
			return NilValue(), err
		}
		if v.Kind() != KindBool {
			return NilValue(), &TypeError{Message: "operand of ! is not bool", Line: line, Col: col}
		}
		return BoolValue(!v.Bool()), nil
	}, nil
}

func (g *codegenVisitor) genCall(n *ast.CallExpr) (evalFn, error) {
	fn, ok := g.reg.Lookup(n.Name)
	if !ok {
		return nil, &TypeError{
			Message: fmt.Sprintf("unknown function %q", n.Name),
			Line:    n.Line, Col: n.Col,
		}
	}

	// Compile argument closures.
	argFns := make([]evalFn, len(n.Args))
	for i, arg := range n.Args {
		af, err := g.generate(arg)
		if err != nil {
			return nil, err
		}
		argFns[i] = af
	}

	return func(ctx context.Context, rt *Runtime) (Value, error) {
		args := make([]Value, len(argFns))
		for i, af := range argFns {
			v, err := af(ctx, rt)
			if err != nil {
				return NilValue(), err
			}
			args[i] = v
		}
		return fn.Impl(ctx, rt, args)
	}, nil
}

func (g *codegenVisitor) genMapIndex(n *ast.MapIndex) (evalFn, error) {
	key := n.Key
	line, col := n.Line, n.Col
	return func(_ context.Context, rt *Runtime) (Value, error) {
		if n.Map == "features" {
			fv, ok := rt.Features[key]
			if !ok {
				return NilValue(), nil
			}
			return featureValueToDSL(fv), nil
		}
		return NilValue(), &TypeError{
			Message: fmt.Sprintf("map %q is not accessible in DSL context", n.Map),
			Line:    line, Col: col,
		}
	}, nil
}

func (g *codegenVisitor) genFieldAccess(n *ast.FieldAccess) (evalFn, error) {
	base, err := g.generate(n.Base)
	if err != nil {
		return nil, err
	}
	fields := n.Fields
	line, col := n.Line, n.Col
	return func(ctx context.Context, rt *Runtime) (Value, error) {
		v, err := base(ctx, rt)
		if err != nil {
			return NilValue(), err
		}
		for _, field := range fields {
			fv, ok := v.Field(field)
			if !ok {
				return NilValue(), &TypeError{
					Message: fmt.Sprintf("field %q not found on object", field),
					Line:    line, Col: col,
				}
			}
			v = fv
		}
		return v, nil
	}, nil
}

func (g *codegenVisitor) genIdent(n *ast.Ident) (evalFn, error) {
	name := n.Name
	line, col := n.Line, n.Col

	// Known top-level request variables.
	switch name {
	case "amount":
		return func(_ context.Context, rt *Runtime) (Value, error) {
			return IntValue(rt.Request.Amount), nil
		}, nil
	case "userID":
		return func(_ context.Context, rt *Runtime) (Value, error) {
			return StringValue(rt.Request.UserID), nil
		}, nil
	case "deviceID":
		return func(_ context.Context, rt *Runtime) (Value, error) {
			return StringValue(rt.Request.DeviceID), nil
		}, nil
	case "ip":
		return func(_ context.Context, rt *Runtime) (Value, error) {
			return StringValue(rt.Request.IP), nil
		}, nil
	case "phone":
		return func(_ context.Context, rt *Runtime) (Value, error) {
			return StringValue(rt.Request.Extra["phone"]), nil
		}, nil
	}
	return nil, &TypeError{
		Message: fmt.Sprintf("unknown identifier %q — use features['key'] for feature values", name),
		Line:    line, Col: col,
	}
}

// ─── Comparison helpers ───────────────────────────────────────────────────────

func compare(op string, l, r Value, line, col int) (bool, error) {
	// String equality/inequality.
	if l.Kind() == KindString && r.Kind() == KindString {
		switch op {
		case "==":
			return l.Str() == r.Str(), nil
		case "!=":
			return l.Str() != r.Str(), nil
		default:
			return false, &TypeError{
				Message: fmt.Sprintf("operator %q not applicable to strings", op),
				Line:    line, Col: col,
			}
		}
	}

	// Bool equality/inequality.
	if l.Kind() == KindBool && r.Kind() == KindBool {
		switch op {
		case "==":
			return l.Bool() == r.Bool(), nil
		case "!=":
			return l.Bool() != r.Bool(), nil
		default:
			return false, &TypeError{
				Message: fmt.Sprintf("operator %q not applicable to booleans", op),
				Line:    line, Col: col,
			}
		}
	}

	// Nil value: treat as numeric zero for comparisons (graceful degradation when
// a feature key is absent). This mirrors the behaviour of most rule engines.
	if l.Kind() == KindNil {
		l = IntValue(0)
	}
	if r.Kind() == KindNil {
		r = IntValue(0)
	}

	// Numeric comparison (int or float, mixed is allowed).
	lf, lok := l.Numeric()
	rf, rok := r.Numeric()
	if !lok || !rok {
		return false, &TypeError{
			Message: fmt.Sprintf("operator %q: incompatible types %d and %d", op, l.Kind(), r.Kind()),
			Line:    line, Col: col,
		}
	}
	switch op {
	case ">":
		return lf > rf, nil
	case "<":
		return lf < rf, nil
	case ">=":
		return lf >= rf, nil
	case "<=":
		return lf <= rf, nil
	case "==":
		return lf == rf, nil
	case "!=":
		return lf != rf, nil
	}
	return false, fmt.Errorf("dsl: unknown operator %q", op)
}
