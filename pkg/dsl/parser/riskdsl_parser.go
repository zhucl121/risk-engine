// Code generated from RiskDSL.g4 by ANTLR 4.13.2. DO NOT EDIT.

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
		"", "", "", "'in'", "'not'", "", "", "", "", "'&&'", "'||'", "'!'",
		"'>'", "'<'", "'>='", "'<='", "'=='", "'!='", "'?'", "':'", "'('", "')'",
		"'['", "']'", "'.'", "','",
	}
	staticData.SymbolicNames = []string{
		"", "BOOL_LIT", "NULL_LIT", "IN", "NOT_KW", "INT_LIT", "FLOAT_LIT",
		"STRING", "ID", "AND", "OR", "NOT", "GT", "LT", "GTE", "LTE", "EQ",
		"NEQ", "QMARK", "COLON", "LPAREN", "RPAREN", "LBRACK", "RBRACK", "DOT",
		"COMMA", "WS",
	}
	staticData.RuleNames = []string{
		"condition", "expr", "argList", "literal",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 26, 84, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 1, 0, 1,
		0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 18, 8, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 4, 1, 28, 8, 1, 11, 1, 12, 1, 29, 1,
		1, 1, 1, 3, 1, 34, 8, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1,
		43, 8, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 5, 1, 64, 8, 1, 10,
		1, 12, 1, 67, 9, 1, 1, 2, 1, 2, 1, 2, 5, 2, 72, 8, 2, 10, 2, 12, 2, 75,
		9, 2, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 3, 3, 82, 8, 3, 1, 3, 0, 1, 2, 4, 0,
		2, 4, 6, 0, 2, 1, 0, 12, 17, 1, 0, 9, 10, 99, 0, 8, 1, 0, 0, 0, 2, 42,
		1, 0, 0, 0, 4, 68, 1, 0, 0, 0, 6, 81, 1, 0, 0, 0, 8, 9, 3, 2, 1, 0, 9,
		10, 5, 0, 0, 1, 10, 1, 1, 0, 0, 0, 11, 12, 6, 1, -1, 0, 12, 13, 5, 11,
		0, 0, 13, 43, 3, 2, 1, 13, 14, 15, 5, 8, 0, 0, 15, 17, 5, 20, 0, 0, 16,
		18, 3, 4, 2, 0, 17, 16, 1, 0, 0, 0, 17, 18, 1, 0, 0, 0, 18, 19, 1, 0, 0,
		0, 19, 43, 5, 21, 0, 0, 20, 21, 5, 8, 0, 0, 21, 22, 5, 22, 0, 0, 22, 23,
		5, 7, 0, 0, 23, 43, 5, 23, 0, 0, 24, 27, 5, 8, 0, 0, 25, 26, 5, 24, 0,
		0, 26, 28, 5, 8, 0, 0, 27, 25, 1, 0, 0, 0, 28, 29, 1, 0, 0, 0, 29, 27,
		1, 0, 0, 0, 29, 30, 1, 0, 0, 0, 30, 43, 1, 0, 0, 0, 31, 33, 5, 22, 0, 0,
		32, 34, 3, 4, 2, 0, 33, 32, 1, 0, 0, 0, 33, 34, 1, 0, 0, 0, 34, 35, 1,
		0, 0, 0, 35, 43, 5, 23, 0, 0, 36, 37, 5, 20, 0, 0, 37, 38, 3, 2, 1, 0,
		38, 39, 5, 21, 0, 0, 39, 43, 1, 0, 0, 0, 40, 43, 3, 6, 3, 0, 41, 43, 5,
		8, 0, 0, 42, 11, 1, 0, 0, 0, 42, 14, 1, 0, 0, 0, 42, 20, 1, 0, 0, 0, 42,
		24, 1, 0, 0, 0, 42, 31, 1, 0, 0, 0, 42, 36, 1, 0, 0, 0, 42, 40, 1, 0, 0,
		0, 42, 41, 1, 0, 0, 0, 43, 65, 1, 0, 0, 0, 44, 45, 10, 12, 0, 0, 45, 46,
		7, 0, 0, 0, 46, 64, 3, 2, 1, 13, 47, 48, 10, 11, 0, 0, 48, 49, 5, 4, 0,
		0, 49, 50, 5, 3, 0, 0, 50, 64, 3, 2, 1, 12, 51, 52, 10, 10, 0, 0, 52, 53,
		5, 3, 0, 0, 53, 64, 3, 2, 1, 11, 54, 55, 10, 9, 0, 0, 55, 56, 7, 1, 0,
		0, 56, 64, 3, 2, 1, 10, 57, 58, 10, 8, 0, 0, 58, 59, 5, 18, 0, 0, 59, 60,
		3, 2, 1, 0, 60, 61, 5, 19, 0, 0, 61, 62, 3, 2, 1, 9, 62, 64, 1, 0, 0, 0,
		63, 44, 1, 0, 0, 0, 63, 47, 1, 0, 0, 0, 63, 51, 1, 0, 0, 0, 63, 54, 1,
		0, 0, 0, 63, 57, 1, 0, 0, 0, 64, 67, 1, 0, 0, 0, 65, 63, 1, 0, 0, 0, 65,
		66, 1, 0, 0, 0, 66, 3, 1, 0, 0, 0, 67, 65, 1, 0, 0, 0, 68, 73, 3, 2, 1,
		0, 69, 70, 5, 25, 0, 0, 70, 72, 3, 2, 1, 0, 71, 69, 1, 0, 0, 0, 72, 75,
		1, 0, 0, 0, 73, 71, 1, 0, 0, 0, 73, 74, 1, 0, 0, 0, 74, 5, 1, 0, 0, 0,
		75, 73, 1, 0, 0, 0, 76, 82, 5, 5, 0, 0, 77, 82, 5, 6, 0, 0, 78, 82, 5,
		7, 0, 0, 79, 82, 5, 1, 0, 0, 80, 82, 5, 2, 0, 0, 81, 76, 1, 0, 0, 0, 81,
		77, 1, 0, 0, 0, 81, 78, 1, 0, 0, 0, 81, 79, 1, 0, 0, 0, 81, 80, 1, 0, 0,
		0, 82, 7, 1, 0, 0, 0, 8, 17, 29, 33, 42, 63, 65, 73, 81,
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
	RiskDSLParserNULL_LIT  = 2
	RiskDSLParserIN        = 3
	RiskDSLParserNOT_KW    = 4
	RiskDSLParserINT_LIT   = 5
	RiskDSLParserFLOAT_LIT = 6
	RiskDSLParserSTRING    = 7
	RiskDSLParserID        = 8
	RiskDSLParserAND       = 9
	RiskDSLParserOR        = 10
	RiskDSLParserNOT       = 11
	RiskDSLParserGT        = 12
	RiskDSLParserLT        = 13
	RiskDSLParserGTE       = 14
	RiskDSLParserLTE       = 15
	RiskDSLParserEQ        = 16
	RiskDSLParserNEQ       = 17
	RiskDSLParserQMARK     = 18
	RiskDSLParserCOLON     = 19
	RiskDSLParserLPAREN    = 20
	RiskDSLParserRPAREN    = 21
	RiskDSLParserLBRACK    = 22
	RiskDSLParserRBRACK    = 23
	RiskDSLParserDOT       = 24
	RiskDSLParserCOMMA     = 25
	RiskDSLParserWS        = 26
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

type InContext struct {
	ExprContext
	left  IExprContext
	right IExprContext
}

func NewInContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *InContext {
	var p = new(InContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *InContext) GetLeft() IExprContext { return s.left }

func (s *InContext) GetRight() IExprContext { return s.right }

func (s *InContext) SetLeft(v IExprContext) { s.left = v }

func (s *InContext) SetRight(v IExprContext) { s.right = v }

func (s *InContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InContext) IN() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserIN, 0)
}

func (s *InContext) AllExpr() []IExprContext {
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

func (s *InContext) Expr(i int) IExprContext {
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

func (s *InContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitIn(s)

	default:
		return t.VisitChildren(s)
	}
}

type TernaryContext struct {
	ExprContext
	cond     IExprContext
	thenExpr IExprContext
	elseExpr IExprContext
}

func NewTernaryContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *TernaryContext {
	var p = new(TernaryContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *TernaryContext) GetCond() IExprContext { return s.cond }

func (s *TernaryContext) GetThenExpr() IExprContext { return s.thenExpr }

func (s *TernaryContext) GetElseExpr() IExprContext { return s.elseExpr }

func (s *TernaryContext) SetCond(v IExprContext) { s.cond = v }

func (s *TernaryContext) SetThenExpr(v IExprContext) { s.thenExpr = v }

func (s *TernaryContext) SetElseExpr(v IExprContext) { s.elseExpr = v }

func (s *TernaryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TernaryContext) QMARK() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserQMARK, 0)
}

func (s *TernaryContext) COLON() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserCOLON, 0)
}

func (s *TernaryContext) AllExpr() []IExprContext {
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

func (s *TernaryContext) Expr(i int) IExprContext {
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

func (s *TernaryContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitTernary(s)

	default:
		return t.VisitChildren(s)
	}
}

type NotInContext struct {
	ExprContext
	left  IExprContext
	right IExprContext
}

func NewNotInContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NotInContext {
	var p = new(NotInContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *NotInContext) GetLeft() IExprContext { return s.left }

func (s *NotInContext) GetRight() IExprContext { return s.right }

func (s *NotInContext) SetLeft(v IExprContext) { s.left = v }

func (s *NotInContext) SetRight(v IExprContext) { s.right = v }

func (s *NotInContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NotInContext) NOT_KW() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserNOT_KW, 0)
}

func (s *NotInContext) IN() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserIN, 0)
}

func (s *NotInContext) AllExpr() []IExprContext {
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

func (s *NotInContext) Expr(i int) IExprContext {
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

func (s *NotInContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitNotIn(s)

	default:
		return t.VisitChildren(s)
	}
}

type ArrayLiteralContext struct {
	ExprContext
}

func NewArrayLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ArrayLiteralContext {
	var p = new(ArrayLiteralContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *ArrayLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArrayLiteralContext) LBRACK() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserLBRACK, 0)
}

func (s *ArrayLiteralContext) RBRACK() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserRBRACK, 0)
}

func (s *ArrayLiteralContext) ArgList() IArgListContext {
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

func (s *ArrayLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitArrayLiteral(s)

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
	p.SetState(42)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext()) {
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

			var _x = p.expr(13)

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

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&5245414) != 0 {
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
		localctx = NewArrayLiteralContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(31)
			p.Match(RiskDSLParserLBRACK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(33)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&5245414) != 0 {
			{
				p.SetState(32)
				p.ArgList()
			}

		}
		{
			p.SetState(35)
			p.Match(RiskDSLParserRBRACK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 6:
		localctx = NewParenContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(36)
			p.Match(RiskDSLParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(37)

			var _x = p.expr(0)

			localctx.(*ParenContext).inner = _x
		}
		{
			p.SetState(38)
			p.Match(RiskDSLParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 7:
		localctx = NewLiteralExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(40)

			var _x = p.Literal()

			localctx.(*LiteralExprContext).lit = _x
		}

	case 8:
		localctx = NewIdentExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(41)

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
	p.SetState(65)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(63)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext()) {
			case 1:
				localctx = NewBinaryCompareContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*BinaryCompareContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, RiskDSLParserRULE_expr)
				p.SetState(44)

				if !(p.Precpred(p.GetParserRuleContext(), 12)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 12)", ""))
					goto errorExit
				}
				{
					p.SetState(45)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*BinaryCompareContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&258048) != 0) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*BinaryCompareContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(46)

					var _x = p.expr(13)

					localctx.(*BinaryCompareContext).right = _x
				}

			case 2:
				localctx = NewNotInContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*NotInContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, RiskDSLParserRULE_expr)
				p.SetState(47)

				if !(p.Precpred(p.GetParserRuleContext(), 11)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 11)", ""))
					goto errorExit
				}
				{
					p.SetState(48)
					p.Match(RiskDSLParserNOT_KW)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(49)
					p.Match(RiskDSLParserIN)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(50)

					var _x = p.expr(12)

					localctx.(*NotInContext).right = _x
				}

			case 3:
				localctx = NewInContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*InContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, RiskDSLParserRULE_expr)
				p.SetState(51)

				if !(p.Precpred(p.GetParserRuleContext(), 10)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 10)", ""))
					goto errorExit
				}
				{
					p.SetState(52)
					p.Match(RiskDSLParserIN)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(53)

					var _x = p.expr(11)

					localctx.(*InContext).right = _x
				}

			case 4:
				localctx = NewBinaryLogicalContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*BinaryLogicalContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, RiskDSLParserRULE_expr)
				p.SetState(54)

				if !(p.Precpred(p.GetParserRuleContext(), 9)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 9)", ""))
					goto errorExit
				}
				{
					p.SetState(55)

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
					p.SetState(56)

					var _x = p.expr(10)

					localctx.(*BinaryLogicalContext).right = _x
				}

			case 5:
				localctx = NewTernaryContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*TernaryContext).cond = _prevctx

				p.PushNewRecursionContext(localctx, _startState, RiskDSLParserRULE_expr)
				p.SetState(57)

				if !(p.Precpred(p.GetParserRuleContext(), 8)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 8)", ""))
					goto errorExit
				}
				{
					p.SetState(58)
					p.Match(RiskDSLParserQMARK)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(59)

					var _x = p.expr(0)

					localctx.(*TernaryContext).thenExpr = _x
				}
				{
					p.SetState(60)
					p.Match(RiskDSLParserCOLON)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(61)

					var _x = p.expr(9)

					localctx.(*TernaryContext).elseExpr = _x
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(67)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext())
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
		p.SetState(68)
		p.expr(0)
	}
	p.SetState(73)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == RiskDSLParserCOMMA {
		{
			p.SetState(69)
			p.Match(RiskDSLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(70)
			p.expr(0)
		}

		p.SetState(75)
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

type NullLiteralContext struct {
	LiteralContext
}

func NewNullLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NullLiteralContext {
	var p = new(NullLiteralContext)

	InitEmptyLiteralContext(&p.LiteralContext)
	p.parser = parser
	p.CopyAll(ctx.(*LiteralContext))

	return p
}

func (s *NullLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NullLiteralContext) NULL_LIT() antlr.TerminalNode {
	return s.GetToken(RiskDSLParserNULL_LIT, 0)
}

func (s *NullLiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case RiskDSLVisitor:
		return t.VisitNullLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *RiskDSLParser) Literal() (localctx ILiteralContext) {
	localctx = NewLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, RiskDSLParserRULE_literal)
	p.SetState(81)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case RiskDSLParserINT_LIT:
		localctx = NewIntLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(76)
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
			p.SetState(77)
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
			p.SetState(78)
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
			p.SetState(79)
			p.Match(RiskDSLParserBOOL_LIT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case RiskDSLParserNULL_LIT:
		localctx = NewNullLiteralContext(p, localctx)
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(80)
			p.Match(RiskDSLParserNULL_LIT)
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
		return p.Precpred(p.GetParserRuleContext(), 12)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 11)

	case 2:
		return p.Precpred(p.GetParserRuleContext(), 10)

	case 3:
		return p.Precpred(p.GetParserRuleContext(), 9)

	case 4:
		return p.Precpred(p.GetParserRuleContext(), 8)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
