// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package dsl

import "fmt"

// SyntaxError is returned when the DSL string cannot be parsed.
type SyntaxError struct {
	Message string
	Line    int
	Col     int
}

func (e *SyntaxError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("dsl syntax error at %d:%d: %s", e.Line, e.Col, e.Message)
	}
	return "dsl syntax error: " + e.Message
}

// TypeError is returned when the DSL expression contains a type mismatch.
type TypeError struct {
	Message string
	Line    int
	Col     int
}

func (e *TypeError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("dsl type error at %d:%d: %s", e.Line, e.Col, e.Message)
	}
	return "dsl type error: " + e.Message
}
