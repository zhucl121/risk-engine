// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package feature

// Standard feature keys used across rules and models.
// Define all keys here to prevent typo divergence between fetchers and consumers.
const (
	// Device features
	KeyDeviceLinkedAccounts7d  = "device.linked_account_count_7d"
	KeyDeviceLinkedAccounts30d = "device.linked_account_count_30d"
	KeyDeviceRiskScore         = "device.risk_score"
	KeyDeviceIsSimulator       = "device.is_simulator"
	KeyDeviceIsRooted          = "device.is_rooted"
	KeyDeviceIsVPN             = "device.is_vpn"

	// User features
	KeyUserRegisterDays        = "user.register_days"
	KeyUserHistoryFraud        = "user.history_fraud_count"
	KeyUserCreditScore         = "user.credit_score"
	KeyUserActiveOrderCount30d = "user.active_order_count_30d"

	// IP features
	KeyIPCountry    = "ip.country"
	KeyIPISDatacenter = "ip.is_datacenter"
	KeyIPRiskScore  = "ip.risk_score"

	// Velocity features (computed by sliding window)
	KeyVelocityPayCount1m   = "velocity.pay_count_1min"
	KeyVelocityPayCount1h   = "velocity.pay_count_1hour"
	KeyVelocityPayCount24h  = "velocity.pay_count_24hour"
	KeyVelocityPromoCount1d = "velocity.promo_claim_count_1d"

	// Session features
	KeySessionDurationSec = "session.duration_sec"
	KeySessionPageCount   = "session.page_count"
)
