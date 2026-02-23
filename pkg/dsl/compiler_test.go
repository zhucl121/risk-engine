// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package dsl_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/riskengine/internal/engine"
	"github.com/yourorg/riskengine/internal/feature"
	"github.com/yourorg/riskengine/pkg/dsl"
	"github.com/yourorg/riskengine/pkg/dsl/builtins"
)

// testHelper is the minimal interface satisfied by both *testing.T and *testing.B.
type testHelper interface {
	Helper()
	Fatalf(string, ...any)
}

func newTestRegistry(tb testHelper) *dsl.FunctionRegistry {
	tb.Helper()
	reg := dsl.NewFunctionRegistry()
	if err := builtins.RegisterAll(reg, builtins.Deps{}); err != nil {
		tb.Fatalf("RegisterAll: %v", err)
	}
	return reg
}


func newRuntime(features map[string]feature.Value, amount int64) *dsl.Runtime {
	rt := dsl.AcquireRuntime()
	fm := make(feature.Map)
	for k, v := range features {
		fm[k] = v
	}
	// amount is now a business field carried via Extra / injected into feature map.
	extra := make(map[string]string)
	if amount != 0 {
		fm["extra.amount"] = feature.Value{Kind: feature.KindInt, IntVal: amount}
		extra["amount"] = fmt.Sprintf("%d", amount)
	}
	rt.Features = fm
	rt.Request = &engine.DecisionRequest{UserID: "u1", IP: "1.2.3.4", Extra: extra}
	return rt
}

// ─── Compile: valid expressions ───────────────────────────────────────────────

func TestCompile_ValidExpressions(t *testing.T) {
	reg := newTestRegistry(t)
	cases := []string{
		"amount > 100",
		"amount >= 0 && amount <= 999999",
		"features['user.register_days'] < 30",
		"features['ip.is_datacenter'] == true && amount > 100000",
		"!features['user.is_verified'] == true",
		"within(amount, 0, 50000)",
		"inList('blacklist.phone', ip)",
		"velocity('pay', '1m') > 5",
		"modelScore('xgb') > 0.8",
		"(amount > 1000 || features['user.register_days'] > 7) && amount < 999999",
	}
	for _, cond := range cases {
		t.Run(cond, func(t *testing.T) {
			prog, err := dsl.Compile(cond, reg)
			require.NoError(t, err, "compile should succeed")
			assert.Equal(t, cond, prog.Source())
		})
	}
}

// ─── Compile: syntax errors ───────────────────────────────────────────────────

func TestCompile_SyntaxErrors(t *testing.T) {
	reg := newTestRegistry(t)
	cases := []struct {
		name      string
		condition string
	}{
		{"empty", ""},
		{"double operator", "amount > > 5"},
		{"unclosed paren", "(amount > 5"},
		{"unclosed string", "features['key"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := dsl.Compile(tc.condition, reg)
			require.Error(t, err)
		})
	}
}

// ─── Compile: unknown function ─────────────────────────────────────────────────

func TestCompile_UnknownFunction(t *testing.T) {
	reg := newTestRegistry(t)
	_, err := dsl.Compile("unknownFunc('x') > 5", reg)
	require.Error(t, err)
	assert.IsType(t, &dsl.TypeError{}, err)
}

// ─── Program.Run: correctness ─────────────────────────────────────────────────

func TestProgram_Run(t *testing.T) {
	reg := newTestRegistry(t)
	cases := []struct {
		name      string
		condition string
		features  map[string]feature.Value
		amount    int64
		want      bool
	}{
		{
			name:      "amount greater than literal",
			condition: "amount > 100",
			amount:    200,
			want:      true,
		},
		{
			name:      "amount less than literal",
			condition: "amount > 100",
			amount:    50,
			want:      false,
		},
		{
			name:      "feature int comparison true",
			condition: "features['user.register_days'] < 30",
			features:  map[string]feature.Value{"user.register_days": {Kind: feature.KindInt, IntVal: 5}},
			want:      true,
		},
		{
			name:      "feature int comparison false",
			condition: "features['user.register_days'] < 30",
			features:  map[string]feature.Value{"user.register_days": {Kind: feature.KindInt, IntVal: 60}},
			want:      false,
		},
		{
			name:      "logical and both true",
			condition: "amount > 10 && amount < 1000",
			amount:    500,
			want:      true,
		},
		{
			name:      "logical and short-circuit false",
			condition: "amount > 10 && amount < 100",
			amount:    500,
			want:      false,
		},
		{
			name:      "logical or first true",
			condition: "amount > 1000 || amount < 100",
			amount:    50,
			want:      true,
		},
		{
			name:      "not operator",
			condition: "!(amount > 1000)",
			amount:    500,
			want:      true,
		},
		{
			name:      "bool feature true",
			condition: "features['ip.is_datacenter'] == true && amount > 100000",
			features:  map[string]feature.Value{"ip.is_datacenter": {Kind: feature.KindBool, BoolVal: true}},
			amount:    200000,
			want:      true,
		},
		{
			name:      "bool feature false",
			condition: "features['ip.is_datacenter'] == true && amount > 100000",
			features:  map[string]feature.Value{"ip.is_datacenter": {Kind: feature.KindBool, BoolVal: false}},
			amount:    200000,
			want:      false,
		},
		{
			name:      "within true",
			condition: "within(amount, 0, 50000)",
			amount:    25000,
			want:      true,
		},
		{
			name:      "within false",
			condition: "within(amount, 0, 50000)",
			amount:    100000,
			want:      false,
		},
		{
			name:      "missing feature returns nil (false comparison)",
			condition: "features['nonexistent'] > 0",
			want:      false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prog, err := dsl.Compile(tc.condition, reg)
			require.NoError(t, err)

			rt := newRuntime(tc.features, tc.amount)
			defer dsl.ReleaseRuntime(rt)

			got, err := prog.Run(context.Background(), rt)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ─── Benchmarks ───────────────────────────────────────────────────────────────

func BenchmarkRunProgram_SimpleCompare(b *testing.B) {
	reg := newTestRegistry(b)
	prog, _ := dsl.Compile("amount > 100", reg)
	rt := newRuntime(nil, 200)
	defer dsl.ReleaseRuntime(rt)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = prog.Run(ctx, rt)
	}
}

func BenchmarkRunProgram_WithFeatures(b *testing.B) {
	reg := newTestRegistry(b)
	prog, _ := dsl.Compile(
		"features['device.linked_account_count_7d'] > 5 && features['user.register_days'] < 30",
		reg,
	)
	features := map[string]feature.Value{
		"device.linked_account_count_7d": {Kind: feature.KindInt, IntVal: 8},
		"user.register_days":             {Kind: feature.KindInt, IntVal: 10},
	}
	rt := newRuntime(features, 0)
	defer dsl.ReleaseRuntime(rt)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = prog.Run(ctx, rt)
	}
}

