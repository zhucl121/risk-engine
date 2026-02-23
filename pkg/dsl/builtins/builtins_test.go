// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package builtins_test

import (
	"context"
	"testing"

	"github.com/yourorg/riskengine/internal/engine"
	"github.com/yourorg/riskengine/pkg/dsl"
	"github.com/yourorg/riskengine/pkg/dsl/builtins"
)

func setupRegistry(t *testing.T) *dsl.FunctionRegistry {
	t.Helper()
	reg := dsl.NewFunctionRegistry()
	if err := builtins.RegisterAll(reg, builtins.Deps{}); err != nil {
		t.Fatalf("RegisterAll: %v", err)
	}
	return reg
}

func eval(t *testing.T, reg *dsl.FunctionRegistry, expr string, features map[string]dsl.Value) (bool, error) {
	t.Helper()
	prog, err := dsl.Compile(expr, reg)
	if err != nil {
		return false, err
	}
	rt := dsl.AcquireRuntime()
	defer dsl.ReleaseRuntime(rt)
	rt.Request = &engine.DecisionRequest{UserID: "u1"}
	return prog.Run(context.Background(), rt)
}

func TestContains(t *testing.T) {
	reg := setupRegistry(t)
	ok, err := eval(t, reg, `contains('hello world', 'world')`, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected true")
	}
}

func TestMatch(t *testing.T) {
	reg := setupRegistry(t)
	cases := []struct {
		expr string
		want bool
	}{
		{`match('abc123', '^[a-z]+[0-9]+$')`, true},
		{`match('ABC', '^[a-z]+$')`, false},
	}
	for _, c := range cases {
		got, err := eval(t, reg, c.expr, nil)
		if err != nil {
			t.Fatalf("%q: %v", c.expr, err)
		}
		if got != c.want {
			t.Errorf("%q: got %v, want %v", c.expr, got, c.want)
		}
	}
}

func TestMathAbsAndMin(t *testing.T) {
	reg := setupRegistry(t)
	rt := dsl.AcquireRuntime()
	defer dsl.ReleaseRuntime(rt)
	rt.Request = &engine.DecisionRequest{}
	ctx := context.Background()

	// abs of a positive float should return itself.
	prog, err := dsl.Compile(`abs(5.0) == 5.0`, reg)
	if err != nil {
		t.Fatalf("compile abs: %v", err)
	}
	if ok, runErr := prog.Run(ctx, rt); runErr != nil || !ok {
		t.Errorf("abs(5.0)==5.0: got %v, err %v", ok, runErr)
	}

	// ceil should round up.
	prog, err = dsl.Compile(`ceil(1.1) == 2.0`, reg)
	if err != nil {
		t.Fatalf("compile ceil: %v", err)
	}
	if ok, runErr := prog.Run(ctx, rt); runErr != nil || !ok {
		t.Errorf("ceil(1.1)==2.0: got %v, err %v", ok, runErr)
	}
}

func TestMinMax(t *testing.T) {
	reg := setupRegistry(t)
	rt := dsl.AcquireRuntime()
	defer dsl.ReleaseRuntime(rt)
	rt.Request = &engine.DecisionRequest{}
	ctx := context.Background()

	prog, _ := dsl.Compile(`min(3.0, 7.0) == 3.0`, reg)
	if ok, err := prog.Run(ctx, rt); err != nil || !ok {
		t.Errorf("min(3,7)==3: got %v, err %v", ok, err)
	}
	prog, _ = dsl.Compile(`max(3.0, 7.0) == 7.0`, reg)
	if ok, err := prog.Run(ctx, rt); err != nil || !ok {
		t.Errorf("max(3,7)==7: got %v, err %v", ok, err)
	}
}

func TestNow(t *testing.T) {
	reg := setupRegistry(t)
	rt := dsl.AcquireRuntime()
	defer dsl.ReleaseRuntime(rt)
	rt.Request = &engine.DecisionRequest{}
	// now() should return a positive Unix timestamp.
	prog, _ := dsl.Compile(`now() > 0`, reg)
	ok, err := prog.Run(context.Background(), rt)
	if err != nil || !ok {
		t.Errorf("now()>0: got %v, err %v", ok, err)
	}
}

func TestToInt(t *testing.T) {
	reg := setupRegistry(t)
	rt := dsl.AcquireRuntime()
	defer dsl.ReleaseRuntime(rt)
	rt.Request = &engine.DecisionRequest{}
	prog, _ := dsl.Compile(`toInt('42') == 42`, reg)
	ok, err := prog.Run(context.Background(), rt)
	if err != nil || !ok {
		t.Errorf("toInt('42')==42: got %v, err %v", ok, err)
	}
}

func TestCoalesce(t *testing.T) {
	reg := setupRegistry(t)
	rt := dsl.AcquireRuntime()
	defer dsl.ReleaseRuntime(rt)
	rt.Request = &engine.DecisionRequest{}
	prog, _ := dsl.Compile(`coalesce('hello') == 'hello'`, reg)
	ok, err := prog.Run(context.Background(), rt)
	if err != nil || !ok {
		t.Errorf("coalesce: got %v, err %v", ok, err)
	}
}

func TestWeekday(t *testing.T) {
	reg := setupRegistry(t)
	rt := dsl.AcquireRuntime()
	defer dsl.ReleaseRuntime(rt)
	rt.Request = &engine.DecisionRequest{}
	prog, _ := dsl.Compile(`weekday() >= 0`, reg)
	ok, err := prog.Run(context.Background(), rt)
	if err != nil || !ok {
		t.Errorf("weekday()>=0: got %v, err %v", ok, err)
	}
}
