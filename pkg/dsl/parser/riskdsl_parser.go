// Code generated from grammar/RiskDSL.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // RiskDSL
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type RiskDSLParser struct {
	*antlr.BaseParser
}

var RiskDSLParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func riskdslParserInit() {
	staticData := &RiskDSLParserStaticData
	staticData.LiteralNames = []string{
		"", "", "", "", "", "", "'&&'", "'||'", "'!'", "'>'", "'<'", "'>='",
		"'<='", "'=='", "'!='", "'('", "')'", "'['", "']'", "'.'", "','",
	}
	staticData.SymbolicNames = []string{
		"", "BOOL_LIT", "INT_LIT", "FLOAT_LIT", "STRING", "ID", "AND", "OR",
		"NOT", "GT", "LT", "GTE", "LTE", "EQ", "NEQ", "LPAREN", "RPAREN", "LBRACK",
		"RBRACK", "DOT", "COMMA", "WS",
	}
	staticData.RuleNames = []string{
		"condition", "expr", "argList", "literal",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 21, 65, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 1, 0, 1,
		0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 18, 8, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 4, 1, 28, 8, 1, 11, 1, 12, 1, 29, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 38, 8, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 5, 1, 46, 8, 1, 10, 1, 12, 1, 49, 9, 1, 1, 2, 1, 2, 1, 2, 5,
		2, 54, 8, 2, 10, 2, 12, 2, 57, 9, 2, 1, 3, 1, 3, 1, 3, 1, 3, 3, 3, 63,
		8, 3, 1, 3, 0, 1, 2, 4, 0, 2, 4, 6, 0, 2, 1, 0, 9, 14, 1, 0, 6, 7, 74,
		0, 8, 1, 0, 0, 0, 2, 37, 1, 0, 0, 0, 4, 50, 1, 0, 0, 0, 6, 62, 1, 0, 0,
		0, 8, 9, 3, 2, 1, 0, 9, 10, 5, 0, 0, 1, 10, 1, 1, 0, 0, 0, 11, 12, 6, 1,
		-1, 0, 12, 13, 5, 8, 0, 0, 13, 38, 3, 2, 1, 9, 14, 15, 5, 5, 0, 0, 15,
		17, 5, 15, 0, 0, 16, 18, 3, 4, 2, 0, 17, 16, 1, 0, 0, 0, 17, 18, 1, 0,
		0, 0, 18, 19, 1, 0, 0, 0, 19, 38, 5, 16, 0, 0, 20, 21, 5, 5, 0, 0, 21,
		22, 5, 17, 0, 0, 22, 23, 5, 4, 0, 0, 23, 38, 5, 18, 0, 0, 24, 27, 5, 5,
		0, 0, 25, 26, 5, 19, 0, 0, 26, 28, 5, 5, 0, 0, 27, 25, 1, 0, 0, 0, 28,
		29, 1, 0, 0, 0, 29, 27, 1, 0, 0, 0, 29, 30, 1, 0, 0, 0, 30, 38, 1, 0, 0,
		0, 31, 32, 5, 15, 0, 0, 32, 33, 3, 2, 1, 0, 33, 34, 5, 16, 0, 0, 34, 38,
		1, 0, 0, 0, 35, 38, 3, 6, 3, 0, 36, 38, 5, 5, 0, 0, 37, 11, 1, 0, 0, 0,
		37, 14, 1, 0, 0, 0, 37, 20, 1, 0, 0, 0, 37, 24, 1, 0, 0, 0, 37, 31, 1,
		0, 0, 0, 37, 35, 1, 0, 0, 0, 37, 36, 1, 0, 0, 0, 38, 47, 1, 0, 0, 0, 39,
		40, 10, 8, 0, 0, 40, 41, 7, 0, 0, 0, 41, 46, 3, 2, 1, 9, 42, 43, 10, 7,
		0, 0, 43, 44, 7, 1, 0, 0, 44, 46, 3, 2, 1, 8, 45, 39, 1, 0, 0, 0, 45, 42,
		1, 0, 0, 0, 46, 49, 1, 0, 0, 0, 47, 45, 1, 0, 0, 0, 47, 48, 1, 0, 0, 0,
		48, 3, 1, 0, 0, 0, 49, 47, 1, 0, 0, 0, 50, 55, 3, 2, 1, 0, 51, 52, 5, 20,
		0, 0, 52, 54, 3, 2, 1, 0, 53, 51, 1, 0, 0, 0, 54, 57, 1, 0, 0, 0, 55, 53,
		1, 0, 0, 0, 55, 56, 1, 0, 0, 0, 56, 5, 1, 0, 0, 0, 57, 55, 1, 0, 0, 0,
		58, 63, 5, 2, 0, 0, 59, 63, 5, 3, 0, 0, 60, 63, 5, 4, 0, 0, 61, 63, 5,
		1, 0, 0, 62, 58, 1, 0, 0, 0, 62, 59, 1, 0, 0, 0, 62, 60, 1, 0, 0, 0, 62,
		61, 1, 0, 0, 0, 63, 7, 1, 0, 0, 0, 7, 17, 29, 37, 45, 47, 55, 62,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// RiskDSLParserInit initializes any static state used to implement RiskDSLParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewRiskDSLParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func RiskDSLParserInit() {
	staticData := &RiskDSLParserStaticData
	staticData.once.Do(riskdslParserInit)
}

// NewRiskDSLParser produces a new parser instance for the optional input antlr.TokenStream.
func NewRiskDSLParser(input antlr.TokenStream) *RiskDSLParser {
	RiskDSLParserInit()
	this := new(RiskDSLParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &RiskDSLParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "RiskDSL.g4"

	return this
}

// RiskDSLParser tokens.
const (
	RiskDSLParserEOF       = antlr.TokenEOF
	RiskDSLParserBOOL_LIT  = 1
	RiskDSLParserINT_LIT   = 2
	RiskDSLParserFLOAT_LIT = 3
	RiskDSLParserSTRING    = 4
	RiskDSLParserID        = 5
	RiskDSLParserAND       = 6
	RiskDSLParserOR        = 7
	RiskDSLParserNOT       = 8
	RiskDSLParserGT        = 9
	RiskDSLParserLT        = 10
	RiskDSLParserGTE       = 11
	RiskDSLParserLTE       = 12
	RiskDSLParserEQ        = 13
	RiskDSLParserNEQ       = 14
	RiskDSLParserLPAREN    = 15
	RiskDSLParserRPAREN    = 16
	RiskDSLParserLBRACK    = 17
	RiskDSLParserRBRACK    = 18
	RiskDSLParserDOT       = 19
	RiskDSLParserCOMMA     = 20
	RiskDSLParserWS        = 21
)

// RiskDSLParser rules.
const (
	RiskDSLParserRULE_condition = 0
	RiskDSLParserRULE_expr      = 1
	RiskDSLParserRULE_argList   = 2
	RiskDSLParserRULE_literal   = 3
)

// IConditionContext is an interface to support dynamic dispatch.
type IConditionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Expr() IExprContext
	EOF() antlr.TerminalNode

	// IsConditionContext differentiates from other interfaces.
	IsConditionContext()
}

type ConditionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyConditionContext() *ConditionContext {
	var p = new(ConditionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = RiskDSLParserRULE_condition
	return p
}

func InitEmptyConditionContext(p *ConditionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = RiskDSLParserRULE_condition
}

func (*ConditionContext) IsConditionContext() {}

func NewConditionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ConditionContext {
	var p = new(ConditionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = RiskDSLParserRULE_condition

	return p
}

func (s *ConditionContext) GetParser() antlr.Parser { return s.parser }

func (s *ConditionContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *ConditionContext) EOF() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserEOF, 0)
}

func (s *ConditionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ConditionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitCondition(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *RiskDSLParser) Condition() (localctx IConditionContext) {
	localctx = NewConditionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, RiskDSLParserRULE_condition)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(8)
		p.expr(0)
	}
	{
		p.SetState(9)
		p.Match(RiskDSLParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExprContext is an interface to support dynamic dispatch.
type IExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsExprContext differentiates from other interfaces.
	IsExprContext()
}

type ExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExprContext() *ExprContext {
	var p = new(ExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = RiskDSLParserRULE_expr
	return p
}

func InitEmptyExprContext(p *ExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = RiskDSLParserRULE_expr
}

func (*ExprContext) IsExprContext() {}

func NewExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExprContext {
	var p = new(ExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = RiskDSLParserRULE_expr

	return p
}

func (s *ExprContext) GetParser() antlr.Parser { return s.parser }

func (s *ExprContext) CopyAll(ctx *ExprContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *ExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type BinaryCompareContext struct {
	ExprContext
	left  IExprContext
	op    antlr.Token
	right IExprContext
}

func NewBinaryCompareContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *BinaryCompareContext {
	var p = new(BinaryCompareContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *BinaryCompareContext) GetOp() antlr.Token { return s.op }

func (s *BinaryCompareContext) SetOp(v antlr.Token) { s.op = v }

func (s *BinaryCompareContext) GetLeft() IExprContext { return s.left }

func (s *BinaryCompareContext) GetRight() IExprContext { return s.right }

func (s *BinaryCompareContext) SetLeft(v IExprContext) { s.left = v }

func (s *BinaryCompareContext) SetRight(v IExprContext) { s.right = v }

func (s *BinaryCompareContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BinaryCompareContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *BinaryCompareContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *BinaryCompareContext) GT() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserGT, 0)
}

func (s *BinaryCompareContext) LT() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserLT, 0)
}

func (s *BinaryCompareContext) GTE() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserGTE, 0)
}

func (s *BinaryCompareContext) LTE() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserLTE, 0)
}

func (s *BinaryCompareContext) EQ() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserEQ, 0)
}

func (s *BinaryCompareContext) NEQ() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserNEQ, 0)
}

func (s *BinaryCompareContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitBinaryCompare(s)

	default:
		return t.VisitChildren(s)
	}
}

type FuncCallContext struct {
	ExprContext
	callee antlr.Token
}

func NewFuncCallContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *FuncCallContext {
	var p = new(FuncCallContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *FuncCallContext) GetCallee() antlr.Token { return s.callee }

func (s *FuncCallContext) SetCallee(v antlr.Token) { s.callee = v }

func (s *FuncCallContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FuncCallContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserLPAREN, 0)
}

func (s *FuncCallContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserRPAREN, 0)
}

func (s *FuncCallContext) ID() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserID, 0)
}

func (s *FuncCallContext) ArgList() IArgListContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArgListContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArgListContext)
}

func (s *FuncCallContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitFuncCall(s)

	default:
		return t.VisitChildren(s)
	}
}

type UnaryNotContext struct {
	ExprContext
	operand IExprContext
}

func NewUnaryNotContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *UnaryNotContext {
	var p = new(UnaryNotContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *UnaryNotContext) GetOperand() IExprContext { return s.operand }

func (s *UnaryNotContext) SetOperand(v IExprContext) { s.operand = v }

func (s *UnaryNotContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UnaryNotContext) NOT() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserNOT, 0)
}

func (s *UnaryNotContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *UnaryNotContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitUnaryNot(s)

	default:
		return t.VisitChildren(s)
	}
}

type IdentExprContext struct {
	ExprContext
	name antlr.Token
}

func NewIdentExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *IdentExprContext {
	var p = new(IdentExprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *IdentExprContext) GetName() antlr.Token { return s.name }

func (s *IdentExprContext) SetName(v antlr.Token) { s.name = v }

func (s *IdentExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IdentExprContext) ID() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserID, 0)
}

func (s *IdentExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitIdentExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type LiteralExprContext struct {
	ExprContext
	lit ILiteralContext
}

func NewLiteralExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LiteralExprContext {
	var p = new(LiteralExprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *LiteralExprContext) GetLit() ILiteralContext { return s.lit }

func (s *LiteralExprContext) SetLit(v ILiteralContext) { s.lit = v }

func (s *LiteralExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralExprContext) Literal() ILiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILiteralContext)
}

func (s *LiteralExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitLiteralExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type MapIndexContext struct {
	ExprContext
	map_ antlr.Token
	key  antlr.Token
}

func NewMapIndexContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MapIndexContext {
	var p = new(MapIndexContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *MapIndexContext) GetMap_() antlr.Token { return s.map_ }

func (s *MapIndexContext) GetKey() antlr.Token { return s.key }

func (s *MapIndexContext) SetMap_(v antlr.Token) { s.map_ = v }

func (s *MapIndexContext) SetKey(v antlr.Token) { s.key = v }

func (s *MapIndexContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MapIndexContext) LBRACK() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserLBRACK, 0)
}

func (s *MapIndexContext) RBRACK() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserRBRACK, 0)
}

func (s *MapIndexContext) ID() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserID, 0)
}

func (s *MapIndexContext) STRING() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserSTRING, 0)
}

func (s *MapIndexContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitMapIndex(s)

	default:
		return t.VisitChildren(s)
	}
}

type FieldAccessContext struct {
	ExprContext
	obj   antlr.Token
	field antlr.Token
}

func NewFieldAccessContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *FieldAccessContext {
	var p = new(FieldAccessContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *FieldAccessContext) GetObj() antlr.Token { return s.obj }

func (s *FieldAccessContext) GetField() antlr.Token { return s.field }

func (s *FieldAccessContext) SetObj(v antlr.Token) { s.obj = v }

func (s *FieldAccessContext) SetField(v antlr.Token) { s.field = v }

func (s *FieldAccessContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FieldAccessContext) AllID() []antlr.TerminalNode {
	return s.GetTokens(RiskDSLParserID)
}

func (s *FieldAccessContext) ID(i int) antlr.TerminalNode {
	return s.GetToken(RiskDSLParserID, i)
}

func (s *FieldAccessContext) AllDOT() []antlr.TerminalNode {
	return s.GetTokens(RiskDSLParserDOT)
}

func (s *FieldAccessContext) DOT(i int) antlr.TerminalNode {
	return s.GetToken(RiskDSLParserDOT, i)
}

func (s *FieldAccessContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitFieldAccess(s)

	default:
		return t.VisitChildren(s)
	}
}

type BinaryLogicalContext struct {
	ExprContext
	left  IExprContext
	op    antlr.Token
	right IExprContext
}

func NewBinaryLogicalContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *BinaryLogicalContext {
	var p = new(BinaryLogicalContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *BinaryLogicalContext) GetOp() antlr.Token { return s.op }

func (s *BinaryLogicalContext) SetOp(v antlr.Token) { s.op = v }

func (s *BinaryLogicalContext) GetLeft() IExprContext { return s.left }

func (s *BinaryLogicalContext) GetRight() IExprContext { return s.right }

func (s *BinaryLogicalContext) SetLeft(v IExprContext) { s.left = v }

func (s *BinaryLogicalContext) SetRight(v IExprContext) { s.right = v }

func (s *BinaryLogicalContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BinaryLogicalContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *BinaryLogicalContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *BinaryLogicalContext) AND() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserAND, 0)
}

func (s *BinaryLogicalContext) OR() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserOR, 0)
}

func (s *BinaryLogicalContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitBinaryLogical(s)

	default:
		return t.VisitChildren(s)
	}
}

type ParenContext struct {
	ExprContext
	inner IExprContext
}

func NewParenContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ParenContext {
	var p = new(ParenContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *ParenContext) GetInner() IExprContext { return s.inner }

func (s *ParenContext) SetInner(v IExprContext) { s.inner = v }

func (s *ParenContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ParenContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserLPAREN, 0)
}

func (s *ParenContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserRPAREN, 0)
}

func (s *ParenContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *ParenContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitParen(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *RiskDSLParser) Expr() (localctx IExprContext) {
	return p.expr(0)
}

func (p *RiskDSLParser) expr(_p int) (localctx IExprContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewExprContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExprContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 2
	p.EnterRecursionRule(localctx, 2, RiskDSLParserRULE_expr, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(37)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext()) {
	case 1:
		localctx = NewUnaryNotContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(12)
			p.Match(RiskDSLParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(13)

			var _x = p.expr(9)

			localctx.(*UnaryNotContext).operand = _x
		}

	case 2:
		localctx = NewFuncCallContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(14)

			var _m = p.Match(RiskDSLParserID)

			localctx.(*FuncCallContext).callee = _m
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(15)
			p.Match(RiskDSLParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(17)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&33086) != 0 {
			{
				p.SetState(16)
				p.ArgList()
			}

		}
		{
			p.SetState(19)
			p.Match(RiskDSLParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		localctx = NewMapIndexContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(20)

			var _m = p.Match(RiskDSLParserID)

			localctx.(*MapIndexContext).map_ = _m
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(21)
			p.Match(RiskDSLParserLBRACK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(22)

			var _m = p.Match(RiskDSLParserSTRING)

			localctx.(*MapIndexContext).key = _m
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(23)
			p.Match(RiskDSLParserRBRACK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		localctx = NewFieldAccessContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(24)

			var _m = p.Match(RiskDSLParserID)

			localctx.(*FieldAccessContext).obj = _m
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(27)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = 1
		for ok := true; ok; ok = _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1:
				{
					p.SetState(25)
					p.Match(RiskDSLParserDOT)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(26)

					var _m = p.Match(RiskDSLParserID)

					localctx.(*FieldAccessContext).field = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			default:
				p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
				goto errorExit
			}

			p.SetState(29)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}

	case 5:
		localctx = NewParenContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(31)
			p.Match(RiskDSLParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(32)

			var _x = p.expr(0)

			localctx.(*ParenContext).inner = _x
		}
		{
			p.SetState(33)
			p.Match(RiskDSLParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 6:
		localctx = NewLiteralExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(35)

			var _x = p.Literal()

			localctx.(*LiteralExprContext).lit = _x
		}

	case 7:
		localctx = NewIdentExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(36)

			var _m = p.Match(RiskDSLParserID)

			localctx.(*IdentExprContext).name = _m
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(47)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(45)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext()) {
			case 1:
				localctx = NewBinaryCompareContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*BinaryCompareContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, RiskDSLParserRULE_expr)
				p.SetState(39)

				if !(p.Precpred(p.GetParserRuleContext(), 8)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 8)", ""))
					goto errorExit
				}
				{
					p.SetState(40)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*BinaryCompareContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&32256) != 0) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*BinaryCompareContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(41)

					var _x = p.expr(9)

					localctx.(*BinaryCompareContext).right = _x
				}

			case 2:
				localctx = NewBinaryLogicalContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*BinaryLogicalContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, RiskDSLParserRULE_expr)
				p.SetState(42)

				if !(p.Precpred(p.GetParserRuleContext(), 7)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 7)", ""))
					goto errorExit
				}
				{
					p.SetState(43)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*BinaryLogicalContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !(_la == RiskDSLParserAND || _la == RiskDSLParserOR) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*BinaryLogicalContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(44)

					var _x = p.expr(8)

					localctx.(*BinaryLogicalContext).right = _x
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(49)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IArgListContext is an interface to support dynamic dispatch.
type IArgListContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllExpr() []IExprContext
	Expr(i int) IExprContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsArgListContext differentiates from other interfaces.
	IsArgListContext()
}

type ArgListContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArgListContext() *ArgListContext {
	var p = new(ArgListContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = RiskDSLParserRULE_argList
	return p
}

func InitEmptyArgListContext(p *ArgListContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = RiskDSLParserRULE_argList
}

func (*ArgListContext) IsArgListContext() {}

func NewArgListContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArgListContext {
	var p = new(ArgListContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = RiskDSLParserRULE_argList

	return p
}

func (s *ArgListContext) GetParser() antlr.Parser { return s.parser }

func (s *ArgListContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *ArgListContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *ArgListContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(RiskDSLParserCOMMA)
}

func (s *ArgListContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(RiskDSLParserCOMMA, i)
}

func (s *ArgListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArgListContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ArgListContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitArgList(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *RiskDSLParser) ArgList() (localctx IArgListContext) {
	localctx = NewArgListContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, RiskDSLParserRULE_argList)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(50)
		p.expr(0)
	}
	p.SetState(55)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == RiskDSLParserCOMMA {
		{
			p.SetState(51)
			p.Match(RiskDSLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(52)
			p.expr(0)
		}

		p.SetState(57)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILiteralContext is an interface to support dynamic dispatch.
type ILiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsLiteralContext differentiates from other interfaces.
	IsLiteralContext()
}

type LiteralContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLiteralContext() *LiteralContext {
	var p = new(LiteralContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = RiskDSLParserRULE_literal
	return p
}

func InitEmptyLiteralContext(p *LiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = RiskDSLParserRULE_literal
}

func (*LiteralContext) IsLiteralContext() {}

func NewLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LiteralContext {
	var p = new(LiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = RiskDSLParserRULE_literal

	return p
}

func (s *LiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *LiteralContext) CopyAll(ctx *LiteralContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *LiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type StringLiteralContext struct {
	LiteralContext
}

func NewStringLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *StringLiteralContext {
	var p = new(StringLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *StringLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StringLiteralContext) STRING() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserSTRING, 0)
}

func (s *StringLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitStringLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

type BoolLiteralContext struct {
	LiteralContext
}

func NewBoolLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *BoolLiteralContext {
	var p = new(BoolLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *BoolLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BoolLiteralContext) BOOL_LIT() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserBOOL_LIT, 0)
}

func (s *BoolLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitBoolLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

type FloatLiteralContext struct {
	LiteralContext
}

func NewFloatLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *FloatLiteralContext {
	var p = new(FloatLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *FloatLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FloatLiteralContext) FLOAT_LIT() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserFLOAT_LIT, 0)
}

func (s *FloatLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitFloatLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

type IntLiteralContext struct {
	LiteralContext
}

func NewIntLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *IntLiteralContext {
	var p = new(IntLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *IntLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IntLiteralContext) INT_LIT() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserINT_LIT, 0)
}

func (s *IntLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitIntLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *RiskDSLParser) Literal() (localctx ILiteralContext) {
	localctx = NewLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, RiskDSLParserRULE_literal)
	p.SetState(62)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case RiskDSLParserINT_LIT:
		localctx = NewIntLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(58)
			p.Match(RiskDSLParserINT_LIT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case RiskDSLParserFLOAT_LIT:
		localctx = NewFloatLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(59)
			p.Match(RiskDSLParserFLOAT_LIT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case RiskDSLParserSTRING:
		localctx = NewStringLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(60)
			p.Match(RiskDSLParserSTRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case RiskDSLParserBOOL_LIT:
		localctx = NewBoolLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(61)
			p.Match(RiskDSLParserBOOL_LIT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

func (p *RiskDSLParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 1:
		var t *ExprContext = nil
		if localctx != nil {
			t = localctx.(*ExprContext)
		}
		return p.Expr_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *RiskDSLParser) Expr_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 8)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 7)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
