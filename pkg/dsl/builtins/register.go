// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package builtins provides the built-in risk function implementations for the
// RiskDSL: inList, velocity, modelScore, geoIP, within.
//
// Usage in cmd/server/main.go:
//
//	reg := dsl.NewFunctionRegistry()
//	err := builtins.RegisterAll(reg, builtins.Deps{
//	    GeoIP: myGeoIPImpl,  // or builtins.StubGeoIPLookup{}
//	})
package builtins

import (
	"fmt"

	"github.com/yourorg/riskengine/pkg/dsl"
)

// Deps holds service dependencies required by builtin functions.
// Fields may be nil; functions will fail-open (return zero value) when nil.
type Deps struct {
	GeoIP GeoIPLookup // used by geoIP(); defaults to StubGeoIPLookup if nil
}

// RegisterAll registers all builtin risk functions into reg.
// Call this once at startup before any calls to dsl.Compile.
func RegisterAll(reg *dsl.FunctionRegistry, deps Deps) error {
	if deps.GeoIP == nil {
		deps.GeoIP = StubGeoIPLookup{}
	}

	registrations := []func(*dsl.FunctionRegistry) error{
		// Core risk functions.
		registerWithin,
		registerInList,
		registerVelocity,
		registerModelScore,
		func(r *dsl.FunctionRegistry) error { return registerGeoIP(r, deps.GeoIP) },

		// String functions: contains, startsWith, endsWith, match, lower, upper, trim, strlen, isEmpty
		registerStrings,
		// Math functions: abs, ceil, floor, round, sqrt, min, max, clamp
		registerMath,
		// Time functions: now, nowMs, daysSince, hoursSince, secondsSince, toUnix, hour, weekday
		registerTime,
		// Type conversion + condition functions: toInt, toFloat, toString, toBool, isNull, coalesce, ifThen
		registerConvert,
	}

	for _, fn := range registrations {
		if err := fn(reg); err != nil {
			return fmt.Errorf("builtins.RegisterAll: %w", err)
		}
	}
	return nil
}
