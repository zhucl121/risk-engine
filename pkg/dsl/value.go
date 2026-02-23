// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package dsl

import "fmt"

// Kind tags the type of a Value, avoiding interface{} boxing in the hot path.
type Kind uint8

const (
	KindNil Kind = iota
	KindBool
	KindInt
	KindFloat
	KindString
	KindObject // structured return value (e.g. GeoInfo)
)

// Value is a tagged union that holds a single DSL runtime value.
// Only one of the concrete fields is valid depending on Kind.
type Value struct {
	kind    Kind
	boolVal bool
	intVal  int64
	fltVal  float64
	strVal  string
	objVal  Object // non-nil only for KindObject
}

// Object is a named-field container used for structured return values
// (e.g. GeoInfo returned by geoIP()). Field access compiles to a map lookup.
type Object map[string]Value

// Typed constructors — avoids accidental field mix-up.

func BoolValue(v bool) Value    { return Value{kind: KindBool, boolVal: v} }
func IntValue(v int64) Value    { return Value{kind: KindInt, intVal: v} }
func FloatValue(v float64) Value { return Value{kind: KindFloat, fltVal: v} }
func StringValue(v string) Value { return Value{kind: KindString, strVal: v} }
func ObjectValue(v Object) Value { return Value{kind: KindObject, objVal: v} }
func NilValue() Value           { return Value{kind: KindNil} }

// Kind returns the value's type tag.
func (v Value) Kind() Kind { return v.kind }

// Bool returns the bool value; panics if Kind != KindBool.
func (v Value) Bool() bool {
	if v.kind != KindBool {
		panic(fmt.Sprintf("dsl: Value.Bool() called on Kind %d", v.kind))
	}
	return v.boolVal
}

// Int returns the int64 value; panics if Kind != KindInt.
func (v Value) Int() int64 {
	if v.kind != KindInt {
		panic(fmt.Sprintf("dsl: Value.Int() called on Kind %d", v.kind))
	}
	return v.intVal
}

// Float returns the float64 value; panics if Kind != KindFloat.
func (v Value) Float() float64 {
	if v.kind != KindFloat {
		panic(fmt.Sprintf("dsl: Value.Float() called on Kind %d", v.kind))
	}
	return v.fltVal
}

// Str returns the string value; panics if Kind != KindString.
func (v Value) Str() string {
	if v.kind != KindString {
		panic(fmt.Sprintf("dsl: Value.Str() called on Kind %d", v.kind))
	}
	return v.strVal
}

// Field looks up a named field on an Object value.
// Returns (value, true) on hit, (NilValue, false) on miss.
func (v Value) Field(name string) (Value, bool) {
	if v.kind != KindObject || v.objVal == nil {
		return NilValue(), false
	}
	fv, ok := v.objVal[name]
	return fv, ok
}

// Numeric returns the value as float64 for comparison purposes.
// Valid for KindInt and KindFloat; returns (0, false) otherwise.
func (v Value) Numeric() (float64, bool) {
	switch v.kind {
	case KindInt:
		return float64(v.intVal), true
	case KindFloat:
		return v.fltVal, true
	}
	return 0, false
}

func (v Value) String() string {
	switch v.kind {
	case KindNil:
		return "<nil>"
	case KindBool:
		if v.boolVal {
			return "true"
		}
		return "false"
	case KindInt:
		return fmt.Sprintf("%d", v.intVal)
	case KindFloat:
		return fmt.Sprintf("%g", v.fltVal)
	case KindString:
		return v.strVal
	case KindObject:
		return fmt.Sprintf("<object %v>", v.objVal)
	}
	return "<unknown>"
}
