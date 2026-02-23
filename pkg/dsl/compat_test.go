// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package dsl_test

// compat_test.go verifies that all existing rule conditions from
// configs/rules/payment_rules.yaml can be compiled by the new DSL Compiler.
// This is the backwards-compatibility gate: if any of these fail, it means
// the Grammar does not cover the existing rule syntax.

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yourorg/riskengine/pkg/dsl"
	"github.com/yourorg/riskengine/pkg/dsl/builtins"
)

// paymentRuleConditions mirrors exactly the condition strings from
// configs/rules/payment_rules.yaml.
var paymentRuleConditions = []struct {
	id        string
	condition string
}{
	{
		"DEVICE_MULTI_ACCOUNT_001",
		"features['device.linked_account_count_7d'] > 5 && features['user.register_days'] < 30",
	},
	{
		"HIGH_VELOCITY_PAY_001",
		"features['velocity.pay_count_1min'] > 5",
	},
	{
		"DATACENTER_IP_LARGE_AMT_001",
		"features['ip.is_datacenter'] == true && amount > 100000",
	},
	{
		"NEW_DEVICE_HIGH_AMT_001",
		"features['user.register_days'] < 7 && amount > 50000",
	},
	{
		"HISTORY_FRAUD_001",
		"features['user.history_fraud_count'] > 0",
	},
}

func TestCompat_PaymentRulesYAML(t *testing.T) {
	reg := dsl.NewFunctionRegistry()
	require.NoError(t, builtins.RegisterAll(reg, builtins.Deps{}))

	for _, tc := range paymentRuleConditions {
		t.Run(tc.id, func(t *testing.T) {
			prog, err := dsl.Compile(tc.condition, reg)
			require.NoError(t, err, "condition from payment_rules.yaml must compile")
			require.NotNil(t, prog)
			require.Equal(t, tc.condition, prog.Source())
		})
	}
}
