// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package ast defines the abstract syntax tree (AST) nodes for the RiskDSL.
// These nodes are the canonical in-memory representation produced by the Compiler
// after parsing; they are independent of the parser backend (hand-written or ANTLR4).
package ast

// Node is the base interface for all AST nodes.
type Node interface {
	// Pos returns the source position (line, col) of the node's first token.
	Pos() (line, col int)
}

// ─── Expression Nodes ────────────────────────────────────────────────────────

// BinaryExpr represents a binary operation: Left OP Right.
// Op is one of: && || > < >= <= == !=
type BinaryExpr struct {
	Op          string
	Left, Right Node
	Line, Col   int
}

func (n *BinaryExpr) Pos() (int, int) { return n.Line, n.Col }

// UnaryExpr represents a unary operation: OP Operand.
// Op is currently only "!".
type UnaryExpr struct {
	Op        string
	Operand   Node
	Line, Col int
}

func (n *UnaryExpr) Pos() (int, int) { return n.Line, n.Col }

// CallExpr represents a function call: Name(Args...).
// The function must be registered in the FunctionRegistry.
type CallExpr struct {
	Name      string
	Args      []Node
	Line, Col int
}

func (n *CallExpr) Pos() (int, int) { return n.Line, n.Col }

// MapIndex represents a map subscript expression: MapName['key'].
// In RiskDSL this is used exclusively for the `features` map, e.g.
// features['user.register_days'].
type MapIndex struct {
	// Map is the identifier of the map variable (e.g. "features").
	Map string
	// Key is the string key, without surrounding quotes.
	Key       string
	Line, Col int
}

func (n *MapIndex) Pos() (int, int) { return n.Line, n.Col }

// FieldAccess represents a dotted field access: Base.Fields[0].Fields[1]...
// Base is typically a CallExpr (e.g. geoIP(ip)) or an Ident.
// Example: geoIP(ip).country  →  FieldAccess{Base: CallExpr{geoIP, [ip]}, Fields: ["country"]}
type FieldAccess struct {
	Base      Node
	Fields    []string
	Line, Col int
}

func (n *FieldAccess) Pos() (int, int) { return n.Line, n.Col }

// Ident is a simple variable reference (e.g. amount, ip, userID).
type Ident struct {
	Name      string
	Line, Col int
}

func (n *Ident) Pos() (int, int) { return n.Line, n.Col }

// ─── Literal Nodes ───────────────────────────────────────────────────────────

// IntLit is an integer literal.
type IntLit struct {
	Val       int64
	Line, Col int
}

func (n *IntLit) Pos() (int, int) { return n.Line, n.Col }

// FloatLit is a floating-point literal.
type FloatLit struct {
	Val       float64
	Line, Col int
}

func (n *FloatLit) Pos() (int, int) { return n.Line, n.Col }

// StringLit is a string literal with quotes already stripped.
type StringLit struct {
	Val       string
	Line, Col int
}

func (n *StringLit) Pos() (int, int) { return n.Line, n.Col }

// BoolLit is a boolean literal (true or false).
type BoolLit struct {
	Val       bool
	Line, Col int
}

func (n *BoolLit) Pos() (int, int) { return n.Line, n.Col }
