// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package dsl provides the RiskEngine self-hosted DSL for rule conditions.
//
// Usage:
//
//	reg := dsl.NewFunctionRegistry()
//	builtins.RegisterAll(reg)  // registers inList, velocity, modelScore, geoIP, within
//
//	prog, err := dsl.Compile("velocity('pay', '1m') > 5 && amount > 10000", reg)
//	if err != nil { /* SyntaxError or TypeError */ }
//
//	rt := dsl.AcquireRuntime()
//	defer dsl.ReleaseRuntime(rt)
//	rt.Features = featureMap
//	rt.Request = req
//	rt.VelocityFunc = velocityService
//
//	hit, err := prog.Run(ctx, rt)
package dsl

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/zhucl121/risk-engine/internal/feature"
	"github.com/zhucl121/risk-engine/pkg/dsl/parser"
)

// Compile parses and compiles a DSL condition string into a reusable Program.
//
// reg must contain all functions referenced in condition; unknown function names
// produce a TypeError at compile time (not at evaluation time).
//
// Compile should only be called during rule loading (hot-reload), never on the
// request hot path.
func Compile(condition string, reg *FunctionRegistry) (*Program, error) {
	if strings.TrimSpace(condition) == "" {
		return nil, &SyntaxError{Message: "condition must not be empty"}
	}

	// Phase 1: ANTLR4 Parse.
	input := antlr.NewInputStream(condition)
	lexer := parser.NewRiskDSLLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewRiskDSLParser(stream)

	// Collect syntax errors.
	errListener := &syntaxErrorListener{}
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(errListener)
	p.RemoveErrorListeners()
	p.AddErrorListener(errListener)

	tree := p.Condition()
	if errListener.err != nil {
		return nil, errListener.err
	}

	// Phase 2: ParseTree → AST (Visitor).
	builder := &astBuilder{}
	raw := tree.Accept(builder)
	res, ok := raw.(visitResult)
	if !ok || res.err != nil {
		msg := "visitor failed"
		if ok && res.err != nil {
			msg = res.err.Error()
		}
		return nil, &SyntaxError{Message: msg}
	}

	// Phase 3: AST → Go closures (codegen).
	gen := &codegenVisitor{reg: reg}
	eval, err := gen.generate(res.node)
	if err != nil {
		return nil, err
	}

	return &Program{src: condition, eval: eval}, nil
}

// syntaxErrorListener collects the first ANTLR4 syntax error.
type syntaxErrorListener struct {
	*antlr.DefaultErrorListener
	err *SyntaxError
}

func (l *syntaxErrorListener) SyntaxError(
	_ antlr.Recognizer,
	_ interface{},
	line, col int,
	msg string,
	_ antlr.RecognitionException,
) {
	if l.err == nil {
		l.err = &SyntaxError{Message: msg, Line: line, Col: col}
	}
}

// featureValueToDSL converts a feature.Value to a DSL Value.
// Centralised here so codegen.go does not import internal/feature directly.
func featureValueToDSL(fv feature.Value) Value {
	switch fv.Kind {
	case feature.KindInt:
		return IntValue(fv.IntVal)
	case feature.KindFloat:
		return FloatValue(fv.FltVal)
	case feature.KindString:
		return StringValue(fv.StrVal)
	case feature.KindBool:
		return BoolValue(fv.BoolVal)
	}
	return NilValue()
}

// DefaultRegistry returns an empty FunctionRegistry.
// Register builtin functions via builtins.RegisterAll(reg, deps) in cmd/server/main.go
// before passing the registry to Compile.
func DefaultRegistry() *FunctionRegistry {
	return NewFunctionRegistry()
}
