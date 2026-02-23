// Code generated from grammar/RiskDSL.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // RiskDSL
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by RiskDSLParser.
type RiskDSLVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by RiskDSLParser#condition.
	VisitCondition(ctx *ConditionContext) interface{}

	// Visit a parse tree produced by RiskDSLParser#BinaryCompare.
	VisitBinaryCompare(ctx *BinaryCompareContext) interface{}

	// Visit a parse tree produced by RiskDSLParser#FuncCall.
	VisitFuncCall(ctx *FuncCallContext) interface{}

	// Visit a parse tree produced by RiskDSLParser#UnaryNot.
	VisitUnaryNot(ctx *UnaryNotContext) interface{}

	// Visit a parse tree produced by RiskDSLParser#IdentExpr.
	VisitIdentExpr(ctx *IdentExprContext) interface{}

	// Visit a parse tree produced by RiskDSLParser#LiteralExpr.
	VisitLiteralExpr(ctx *LiteralExprContext) interface{}

	// Visit a parse tree produced by RiskDSLParser#MapIndex.
	VisitMapIndex(ctx *MapIndexContext) interface{}

	// Visit a parse tree produced by RiskDSLParser#FieldAccess.
	VisitFieldAccess(ctx *FieldAccessContext) interface{}

	// Visit a parse tree produced by RiskDSLParser#BinaryLogical.
	VisitBinaryLogical(ctx *BinaryLogicalContext) interface{}

	// Visit a parse tree produced by RiskDSLParser#Paren.
	VisitParen(ctx *ParenContext) interface{}

	// Visit a parse tree produced by RiskDSLParser#argList.
	VisitArgList(ctx *ArgListContext) interface{}

	// Visit a parse tree produced by RiskDSLParser#IntLiteral.
	VisitIntLiteral(ctx *IntLiteralContext) interface{}

	// Visit a parse tree produced by RiskDSLParser#FloatLiteral.
	VisitFloatLiteral(ctx *FloatLiteralContext) interface{}

	// Visit a parse tree produced by RiskDSLParser#StringLiteral.
	VisitStringLiteral(ctx *StringLiteralContext) interface{}

	// Visit a parse tree produced by RiskDSLParser#BoolLiteral.
	VisitBoolLiteral(ctx *BoolLiteralContext) interface{}
}
