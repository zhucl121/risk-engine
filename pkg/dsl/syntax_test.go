// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package dsl_test

import (
	"context"
	"testing"

	"github.com/yourorg/riskengine/internal/engine"
	"github.com/yourorg/riskengine/pkg/dsl"
	"github.com/yourorg/riskengine/pkg/dsl/builtins"
)

func makeReg(t *testing.T) *dsl.FunctionRegistry {
	t.Helper()
	reg := dsl.NewFunctionRegistry()
	if err := builtins.RegisterAll(reg, builtins.Deps{}); err != nil {
		t.Fatalf("RegisterAll: %v", err)
	}
	return reg
}

func runDSL(t *testing.T, reg *dsl.FunctionRegistry, expr string, amount int64) bool {
	t.Helper()
	prog, err := dsl.Compile(expr, reg)
	if err != nil {
		t.Fatalf("compile %q: %v", expr, err)
	}
	rt := dsl.AcquireRuntime()
	defer dsl.ReleaseRuntime(rt)
	rt.Request = &engine.DecisionRequest{Amount: amount}
	got, err := prog.Run(context.Background(), rt)
	if err != nil {
		t.Fatalf("run %q: %v", expr, err)
	}
	return got
}

func TestInOperator(t *testing.T) {
	reg := makeReg(t)
	tests := []struct {
		expr string
		want bool
	}{
		// amount in [100, 200, 500]
		{`amount in [100, 200, 500]`, true},   // amount=100 → hit
		{`amount in [200, 500, 900]`, false},  // amount=100 → miss
		// not in
		{`amount not in [1, 2, 3]`, true},     // 100 not in [1,2,3]
		{`amount not in [100, 200]`, false},   // 100 is in [100,200]
	}
	for _, tc := range tests {
		got := runDSL(t, reg, tc.expr, 100)
		if got != tc.want {
			t.Errorf("%q (amount=100): got %v, want %v", tc.expr, got, tc.want)
		}
	}
}

func TestTernaryOperator(t *testing.T) {
	reg := makeReg(t)
	// Ternary returns string; compare with == to get bool.
	prog, err := dsl.Compile(`amount > 50 ? 'high' : 'low'`, reg)
	if err != nil {
		// Ternary evaluates to a non-bool — expect compile success but run type error.
		t.Fatalf("compile ternary: %v", err)
	}
	rt := dsl.AcquireRuntime()
	defer dsl.ReleaseRuntime(rt)
	rt.Request = &engine.DecisionRequest{Amount: 100}
	// Run will fail because root must be bool; verify the error is informative.
	_, runErr := prog.Run(context.Background(), rt)
	if runErr == nil {
		t.Fatal("expected type error for non-bool ternary at top level")
	}

	// Ternary used inside a comparison → top-level is bool.
	prog2, err := dsl.Compile(`(amount > 50 ? amount : 0) == amount`, reg)
	if err != nil {
		t.Fatalf("compile ternary-in-compare: %v", err)
	}
	ok, runErr2 := prog2.Run(context.Background(), rt)
	if runErr2 != nil {
		t.Fatalf("run ternary-in-compare: %v", runErr2)
	}
	if !ok {
		t.Error("ternary-in-compare: expected true (amount=100 > 50 → amount == amount)")
	}
}

func TestInWithStrings(t *testing.T) {
	reg := makeReg(t)
	prog, err := dsl.Compile(`ip in ['1.2.3.4', '5.6.7.8', '9.10.11.12']`, reg)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	rt := dsl.AcquireRuntime()
	defer dsl.ReleaseRuntime(rt)
	rt.Request = &engine.DecisionRequest{IP: "1.2.3.4"}
	ok, err := prog.Run(context.Background(), rt)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !ok {
		t.Error("expected ip in list to be true")
	}
}
