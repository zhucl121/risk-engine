// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"

	"github.com/yourorg/riskengine/pkg/dsl"
)

// GeoIPLookup is the interface for IP geolocation. Implementations are injected
// at startup; the default is a stub that returns an empty GeoInfo.
type GeoIPLookup interface {
	Lookup(ctx context.Context, ip string) (GeoInfo, error)
}

// GeoInfo holds geolocation results accessible as fields in DSL (geoIP(ip).country).
type GeoInfo struct {
	Country string
	ISP     string
	ASN     string
	IsProxy bool
}

// geoInfoToObject converts GeoInfo to a dsl.Object for field access.
func geoInfoToObject(g GeoInfo) dsl.Object {
	return dsl.Object{
		"country": dsl.StringValue(g.Country),
		"isp":     dsl.StringValue(g.ISP),
		"asn":     dsl.StringValue(g.ASN),
		"isProxy": dsl.BoolValue(g.IsProxy),
	}
}

// geoIPFunc creates the geoIP(ip) DSL function bound to the given lookup impl.
func geoIPFunc(lookup GeoIPLookup) func(context.Context, *dsl.Runtime, []dsl.Value) (dsl.Value, error) {
	return func(ctx context.Context, _ *dsl.Runtime, args []dsl.Value) (dsl.Value, error) {
		if len(args) != 1 {
			return dsl.NilValue(), fmt.Errorf("geoIP: expected 1 arg, got %d", len(args))
		}
		if args[0].Kind() != dsl.KindString {
			return dsl.NilValue(), fmt.Errorf("geoIP: arg 1 (ip) must be string")
		}
		info, err := lookup.Lookup(ctx, args[0].Str())
		if err != nil {
			// Fail-open: return empty GeoInfo so the rule can still evaluate.
			return dsl.ObjectValue(geoInfoToObject(GeoInfo{})), nil
		}
		return dsl.ObjectValue(geoInfoToObject(info)), nil
	}
}

// StubGeoIPLookup returns an empty GeoInfo for every IP. Used in tests and
// when no real GeoIP service is configured.
type StubGeoIPLookup struct{}

func (StubGeoIPLookup) Lookup(_ context.Context, _ string) (GeoInfo, error) {
	return GeoInfo{Country: "CN"}, nil
}

func registerGeoIP(reg *dsl.FunctionRegistry, lookup GeoIPLookup) error {
	return reg.Register(dsl.FuncDef{
		Name:       "geoIP",
		Args:       []dsl.ArgKind{dsl.ArgKindString},
		ReturnKind: dsl.KindObject,
		Impl:       geoIPFunc(lookup),
	})
}
