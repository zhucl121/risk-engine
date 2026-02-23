// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package dsl

// visitor.go — converts an ANTLR4 ParseTree into the dsl/ast Node tree.
// This decouples the rest of the DSL engine from the ANTLR4 API so that
// the hand-written parser (or any future backend) can be swapped in without
// changing the compiler or codegen layers.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/yourorg/riskengine/pkg/dsl/ast"
	"github.com/yourorg/riskengine/pkg/dsl/parser"
)

// astBuilder implements parser.RiskDSLVisitor and produces ast.Node values.
// Each Visit method returns an ast.Node or error (wrapped in a result struct).
type astBuilder struct {
	parser.BaseRiskDSLVisitor
}

type visitResult struct {
	node ast.Node
	err  error
}

func (b *astBuilder) visit(tree antlr.ParseTree) (ast.Node, error) {
	raw := tree.Accept(b)
	if raw == nil {
		return nil, fmt.Errorf("dsl: visitor returned nil for %T", tree)
	}
	res, ok := raw.(visitResult)
	if !ok {
		return nil, fmt.Errorf("dsl: unexpected visitor result type %T", raw)
	}
	return res.node, res.err
}

func ok(node ast.Node) visitResult   { return visitResult{node: node} }
func fail(err error) visitResult     { return visitResult{err: err} }
func failf(f string, a ...any) visitResult { return visitResult{err: fmt.Errorf(f, a...)} }

// VisitCondition: condition → expr EOF
func (b *astBuilder) VisitCondition(ctx *parser.ConditionContext) interface{} {
	return ctx.Expr().Accept(b)
}

// VisitBinaryLogical: expr ('&&'|'||') expr
func (b *astBuilder) VisitBinaryLogical(ctx *parser.BinaryLogicalContext) interface{} {
	left, err := b.visit(ctx.GetLeft())
	if err != nil {
		return fail(err)
	}
	right, err := b.visit(ctx.GetRight())
	if err != nil {
		return fail(err)
	}
	op := ctx.GetOp().GetText()
	line := ctx.GetOp().GetLine()
	col := ctx.GetOp().GetColumn()
	return ok(&ast.BinaryExpr{Op: op, Left: left, Right: right, Line: line, Col: col})
}

// VisitBinaryCompare: expr OP expr  (>, <, >=, <=, ==, !=)
func (b *astBuilder) VisitBinaryCompare(ctx *parser.BinaryCompareContext) interface{} {
	left, err := b.visit(ctx.GetLeft())
	if err != nil {
		return fail(err)
	}
	right, err := b.visit(ctx.GetRight())
	if err != nil {
		return fail(err)
	}
	op := ctx.GetOp().GetText()
	line := ctx.GetOp().GetLine()
	col := ctx.GetOp().GetColumn()
	return ok(&ast.BinaryExpr{Op: op, Left: left, Right: right, Line: line, Col: col})
}

// VisitUnaryNot: '!' expr
func (b *astBuilder) VisitUnaryNot(ctx *parser.UnaryNotContext) interface{} {
	operand, err := b.visit(ctx.GetOperand())
	if err != nil {
		return fail(err)
	}
	line := ctx.GetStart().GetLine()
	col := ctx.GetStart().GetColumn()
	return ok(&ast.UnaryExpr{Op: "!", Operand: operand, Line: line, Col: col})
}

// VisitFuncCall: ID '(' argList? ')'
func (b *astBuilder) VisitFuncCall(ctx *parser.FuncCallContext) interface{} {
	name := ctx.GetCallee().GetText()
	line := ctx.GetStart().GetLine()
	col := ctx.GetStart().GetColumn()

	var args []ast.Node
	if al := ctx.ArgList(); al != nil {
		raw := al.Accept(b)
		res, ok2 := raw.(visitResult)
		if !ok2 || res.err != nil {
			if ok2 {
				return fail(res.err)
			}
			return failf("dsl: argList visit failed for %s", name)
		}
		// argList returns a slice wrapped in a special node; unwrap here.
		if list, islist := res.node.(*argListNode); islist {
			args = list.nodes
		}
	}
	return ok(&ast.CallExpr{Name: name, Args: args, Line: line, Col: col})
}

// VisitArgList: expr (',' expr)*  — returns a synthetic argListNode
func (b *astBuilder) VisitArgList(ctx *parser.ArgListContext) interface{} {
	var nodes []ast.Node
	for _, exprCtx := range ctx.AllExpr() {
		node, err := b.visit(exprCtx)
		if err != nil {
			return fail(err)
		}
		nodes = append(nodes, node)
	}
	return visitResult{node: &argListNode{nodes: nodes}}
}

// argListNode is a synthetic AST node used only to transport the argument list
// from VisitArgList back to VisitFuncCall. It is never emitted to the compiler.
type argListNode struct{ nodes []ast.Node }

func (n *argListNode) Pos() (int, int) { return 0, 0 }

// VisitMapIndex: ID '[' STRING ']'
func (b *astBuilder) VisitMapIndex(ctx *parser.MapIndexContext) interface{} {
	mapName := ctx.GetMap_().GetText()
	key := stripDSLQuotes(ctx.GetKey().GetText())
	line := ctx.GetStart().GetLine()
	col := ctx.GetStart().GetColumn()
	return ok(&ast.MapIndex{Map: mapName, Key: key, Line: line, Col: col})
}

// VisitFieldAccess: ID ('.' ID)+
func (b *astBuilder) VisitFieldAccess(ctx *parser.FieldAccessContext) interface{} {
	line := ctx.GetStart().GetLine()
	col := ctx.GetStart().GetColumn()

	// The grammar produces: obj=ID ('.' field=ID)+
	// GetObj() is the base identifier.
	baseIdent := &ast.Ident{Name: ctx.GetObj().GetText(), Line: line, Col: col}

	var fields []string
	for _, f := range ctx.AllID() {
		// AllID includes the base obj token and the field tokens.
		// Skip the first one (the obj) — already captured above.
		if fields != nil || f.GetText() != ctx.GetObj().GetText() {
			fields = append(fields, f.GetText())
		}
	}
	// If the base is actually a FuncCall (grammar doesn't directly support that
	// in FieldAccess — handled at postfix level), base stays as Ident.
	return ok(&ast.FieldAccess{Base: baseIdent, Fields: fields, Line: line, Col: col})
}

// VisitParen: '(' expr ')'
func (b *astBuilder) VisitParen(ctx *parser.ParenContext) interface{} {
	return ctx.GetInner().Accept(b)
}

// VisitIdentExpr: ID
func (b *astBuilder) VisitIdentExpr(ctx *parser.IdentExprContext) interface{} {
	name := ctx.GetName().GetText()
	line := ctx.GetStart().GetLine()
	col := ctx.GetStart().GetColumn()
	return ok(&ast.Ident{Name: name, Line: line, Col: col})
}

// VisitLiteralExpr: delegates to the literal sub-rule visitor.
func (b *astBuilder) VisitLiteralExpr(ctx *parser.LiteralExprContext) interface{} {
	return ctx.GetLit().Accept(b)
}

// VisitIntLiteral: INT_LIT
func (b *astBuilder) VisitIntLiteral(ctx *parser.IntLiteralContext) interface{} {
	txt := ctx.GetStart().GetText()
	v, err := strconv.ParseInt(txt, 10, 64)
	if err != nil {
		return failf("dsl: invalid int literal %q", txt)
	}
	return ok(&ast.IntLit{Val: v, Line: ctx.GetStart().GetLine(), Col: ctx.GetStart().GetColumn()})
}

// VisitFloatLiteral: FLOAT_LIT
func (b *astBuilder) VisitFloatLiteral(ctx *parser.FloatLiteralContext) interface{} {
	txt := ctx.GetStart().GetText()
	v, err := strconv.ParseFloat(txt, 64)
	if err != nil {
		return failf("dsl: invalid float literal %q", txt)
	}
	return ok(&ast.FloatLit{Val: v, Line: ctx.GetStart().GetLine(), Col: ctx.GetStart().GetColumn()})
}

// VisitStringLiteral: STRING
func (b *astBuilder) VisitStringLiteral(ctx *parser.StringLiteralContext) interface{} {
	txt := ctx.GetStart().GetText()
	return ok(&ast.StringLit{
		Val:  stripDSLQuotes(txt),
		Line: ctx.GetStart().GetLine(),
		Col:  ctx.GetStart().GetColumn(),
	})
}

// VisitBoolLiteral: BOOL_LIT
func (b *astBuilder) VisitBoolLiteral(ctx *parser.BoolLiteralContext) interface{} {
	return ok(&ast.BoolLit{
		Val:  ctx.GetStart().GetText() == "true",
		Line: ctx.GetStart().GetLine(),
		Col:  ctx.GetStart().GetColumn(),
	})
}

// VisitNullLiteral: NULL_LIT → NilValue
func (b *astBuilder) VisitNullLiteral(ctx *parser.NullLiteralContext) interface{} {
	line := ctx.GetStart().GetLine()
	col := ctx.GetStart().GetColumn()
	// Represent null as a special NilLit node (reuse IntLit{0} as nil sentinel).
	// We use a StringLit with a sentinel so codegen can emit NilValue().
	_ = line
	_ = col
	return ok(&ast.IntLit{Val: 0, Line: line, Col: col}) // treated as KindNil in genIdent
}

// VisitTernary: cond ? then : else
func (b *astBuilder) VisitTernary(ctx *parser.TernaryContext) interface{} {
	cond, err := b.visit(ctx.GetCond())
	if err != nil {
		return fail(err)
	}
	thenExpr, err := b.visit(ctx.GetThenExpr())
	if err != nil {
		return fail(err)
	}
	elseExpr, err := b.visit(ctx.GetElseExpr())
	if err != nil {
		return fail(err)
	}
	line := ctx.GetStart().GetLine()
	col := ctx.GetStart().GetColumn()
	return ok(&ast.TernaryExpr{
		Condition: cond,
		Then:      thenExpr,
		Else:      elseExpr,
		Line:      line, Col: col,
	})
}

// VisitIn: left in right
func (b *astBuilder) VisitIn(ctx *parser.InContext) interface{} {
	left, err := b.visit(ctx.GetLeft())
	if err != nil {
		return fail(err)
	}
	right, err := b.visit(ctx.GetRight())
	if err != nil {
		return fail(err)
	}
	line := ctx.GetStart().GetLine()
	col := ctx.GetStart().GetColumn()
	return ok(&ast.InExpr{Value: left, Array: right, Negated: false, Line: line, Col: col})
}

// VisitNotIn: left not in right
func (b *astBuilder) VisitNotIn(ctx *parser.NotInContext) interface{} {
	left, err := b.visit(ctx.GetLeft())
	if err != nil {
		return fail(err)
	}
	right, err := b.visit(ctx.GetRight())
	if err != nil {
		return fail(err)
	}
	line := ctx.GetStart().GetLine()
	col := ctx.GetStart().GetColumn()
	return ok(&ast.InExpr{Value: left, Array: right, Negated: true, Line: line, Col: col})
}

// VisitArrayLiteral: '[' argList? ']'
func (b *astBuilder) VisitArrayLiteral(ctx *parser.ArrayLiteralContext) interface{} {
	line := ctx.GetStart().GetLine()
	col := ctx.GetStart().GetColumn()

	var elems []ast.Node
	if al := ctx.ArgList(); al != nil {
		raw := al.Accept(b)
		res, ok2 := raw.(visitResult)
		if !ok2 || res.err != nil {
			if ok2 {
				return fail(res.err)
			}
			return failf("dsl: argList visit failed for array literal")
		}
		if list, isList := res.node.(*argListNode); isList {
			elems = list.nodes
		}
	}
	return ok(&ast.ArrayLit{Elems: elems, Line: line, Col: col})
}

// stripDSLQuotes removes surrounding single or double quotes from a DSL string token.
func stripDSLQuotes(s string) string {
	if len(s) >= 2 && (s[0] == '\'' || s[0] == '"') {
		s = s[1 : len(s)-1]
	}
	return strings.ReplaceAll(s, "\\'", "'")
}
