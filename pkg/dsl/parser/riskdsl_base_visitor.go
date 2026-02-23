// Code generated from grammar/RiskDSL.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // RiskDSL
import "github.com/antlr4-go/antlr/v4"

type BaseRiskDSLVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseRiskDSLVisitor) VisitCondition(ctx *ConditionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseRiskDSLVisitor) VisitBinaryCompare(ctx *BinaryCompareContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseRiskDSLVisitor) VisitFuncCall(ctx *FuncCallContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseRiskDSLVisitor) VisitUnaryNot(ctx *UnaryNotContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseRiskDSLVisitor) VisitIdentExpr(ctx *IdentExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseRiskDSLVisitor) VisitLiteralExpr(ctx *LiteralExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseRiskDSLVisitor) VisitMapIndex(ctx *MapIndexContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseRiskDSLVisitor) VisitFieldAccess(ctx *FieldAccessContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseRiskDSLVisitor) VisitBinaryLogical(ctx *BinaryLogicalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseRiskDSLVisitor) VisitParen(ctx *ParenContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseRiskDSLVisitor) VisitArgList(ctx *ArgListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseRiskDSLVisitor) VisitIntLiteral(ctx *IntLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseRiskDSLVisitor) VisitFloatLiteral(ctx *FloatLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseRiskDSLVisitor) VisitStringLiteral(ctx *StringLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseRiskDSLVisitor) VisitBoolLiteral(ctx *BoolLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}
