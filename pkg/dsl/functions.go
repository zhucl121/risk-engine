// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package dsl

import (
	"context"
	"fmt"
)

// ArgKind describes the expected type of a function argument.
type ArgKind uint8

const (
	ArgKindAny ArgKind = iota
	ArgKindBool
	ArgKindInt
	ArgKindFloat
	ArgKindString
	ArgKindNumber // int or float
)

// FuncDef defines a callable DSL function: its signature and Go implementation.
type FuncDef struct {
	// Name is the identifier used in DSL expressions (e.g. "inList").
	Name string
	// Args describes the expected argument kinds for static type checking.
	// Length must match the number of arguments; use ArgKindAny to skip checking.
	Args []ArgKind
	// ReturnKind is the Kind of the value returned by this function.
	ReturnKind Kind
	// Impl is the Go function called at evaluation time.
	// ctx is passed from Program.Run; rt gives access to services.
	Impl func(ctx context.Context, rt *Runtime, args []Value) (Value, error)
}

// FunctionRegistry holds all DSL functions available during compilation and evaluation.
// It is built once at startup and shared across all compilations (read-only after init).
type FunctionRegistry struct {
	funcs map[string]FuncDef
}

// NewFunctionRegistry creates an empty registry.
func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{funcs: make(map[string]FuncDef)}
}

// Register adds a FuncDef to the registry.
// Returns an error if a function with the same name is already registered.
func (r *FunctionRegistry) Register(fn FuncDef) error {
	if _, exists := r.funcs[fn.Name]; exists {
		return fmt.Errorf("dsl: function %q already registered", fn.Name)
	}
	r.funcs[fn.Name] = fn
	return nil
}

// MustRegister calls Register and panics on error (intended for init-time use).
func (r *FunctionRegistry) MustRegister(fn FuncDef) {
	if err := r.Register(fn); err != nil {
		panic(err)
	}
}

// Lookup returns the FuncDef for name, or (zero, false) if not found.
func (r *FunctionRegistry) Lookup(name string) (FuncDef, bool) {
	fn, ok := r.funcs[name]
	return fn, ok
}

// FuncImpl is the Go implementation type for a DSL function.
// It is a convenience alias for the Impl field of FuncDef.
type FuncImpl = func(ctx context.Context, rt *Runtime, args []Value) (Value, error)

// RegisterFunc registers a function by name with any-typed arguments and any return kind.
// This is a convenience wrapper around Register for functions that do not need
// static argument-type checking.
func (r *FunctionRegistry) RegisterFunc(name string, impl FuncImpl) error {
	return r.Register(FuncDef{
		Name:       name,
		Args:       nil, // no static type check
		ReturnKind: KindNil, // unknown at compile time
		Impl:       impl,
	})
}
