// Code generated from RiskDSL.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"sync"
	"unicode"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = sync.Once{}
var _ = unicode.IsLetter

type RiskDSLLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var RiskDSLLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func riskdsllexerLexerInit() {
	staticData := &RiskDSLLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
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
		"BOOL_LIT", "NULL_LIT", "IN", "NOT_KW", "INT_LIT", "FLOAT_LIT", "STRING",
		"ID", "AND", "OR", "NOT", "GT", "LT", "GTE", "LTE", "EQ", "NEQ", "QMARK",
		"COLON", "LPAREN", "RPAREN", "LBRACK", "RBRACK", "DOT", "COMMA", "WS",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 26, 172, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15,
		7, 15, 2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7,
		20, 2, 21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25,
		1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 3, 0, 63, 8, 0, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 72, 8, 1, 1, 2, 1, 2, 1, 2,
		1, 3, 1, 3, 1, 3, 1, 3, 1, 4, 4, 4, 82, 8, 4, 11, 4, 12, 4, 83, 1, 5, 4,
		5, 87, 8, 5, 11, 5, 12, 5, 88, 1, 5, 1, 5, 4, 5, 93, 8, 5, 11, 5, 12, 5,
		94, 1, 6, 1, 6, 1, 6, 1, 6, 5, 6, 101, 8, 6, 10, 6, 12, 6, 104, 9, 6, 1,
		6, 1, 6, 1, 6, 1, 6, 1, 6, 5, 6, 111, 8, 6, 10, 6, 12, 6, 114, 9, 6, 1,
		6, 3, 6, 117, 8, 6, 1, 7, 1, 7, 5, 7, 121, 8, 7, 10, 7, 12, 7, 124, 9,
		7, 1, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1, 9, 1, 10, 1, 10, 1, 11, 1, 11, 1, 12,
		1, 12, 1, 13, 1, 13, 1, 13, 1, 14, 1, 14, 1, 14, 1, 15, 1, 15, 1, 15, 1,
		16, 1, 16, 1, 16, 1, 17, 1, 17, 1, 18, 1, 18, 1, 19, 1, 19, 1, 20, 1, 20,
		1, 21, 1, 21, 1, 22, 1, 22, 1, 23, 1, 23, 1, 24, 1, 24, 1, 25, 4, 25, 167,
		8, 25, 11, 25, 12, 25, 168, 1, 25, 1, 25, 0, 0, 26, 1, 1, 3, 2, 5, 3, 7,
		4, 9, 5, 11, 6, 13, 7, 15, 8, 17, 9, 19, 10, 21, 11, 23, 12, 25, 13, 27,
		14, 29, 15, 31, 16, 33, 17, 35, 18, 37, 19, 39, 20, 41, 21, 43, 22, 45,
		23, 47, 24, 49, 25, 51, 26, 1, 0, 6, 1, 0, 48, 57, 1, 0, 39, 39, 1, 0,
		34, 34, 3, 0, 65, 90, 95, 95, 97, 122, 4, 0, 48, 57, 65, 90, 95, 95, 97,
		122, 3, 0, 9, 10, 13, 13, 32, 32, 183, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0,
		0, 0, 5, 1, 0, 0, 0, 0, 7, 1, 0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0,
		0, 0, 13, 1, 0, 0, 0, 0, 15, 1, 0, 0, 0, 0, 17, 1, 0, 0, 0, 0, 19, 1, 0,
		0, 0, 0, 21, 1, 0, 0, 0, 0, 23, 1, 0, 0, 0, 0, 25, 1, 0, 0, 0, 0, 27, 1,
		0, 0, 0, 0, 29, 1, 0, 0, 0, 0, 31, 1, 0, 0, 0, 0, 33, 1, 0, 0, 0, 0, 35,
		1, 0, 0, 0, 0, 37, 1, 0, 0, 0, 0, 39, 1, 0, 0, 0, 0, 41, 1, 0, 0, 0, 0,
		43, 1, 0, 0, 0, 0, 45, 1, 0, 0, 0, 0, 47, 1, 0, 0, 0, 0, 49, 1, 0, 0, 0,
		0, 51, 1, 0, 0, 0, 1, 62, 1, 0, 0, 0, 3, 71, 1, 0, 0, 0, 5, 73, 1, 0, 0,
		0, 7, 76, 1, 0, 0, 0, 9, 81, 1, 0, 0, 0, 11, 86, 1, 0, 0, 0, 13, 116, 1,
		0, 0, 0, 15, 118, 1, 0, 0, 0, 17, 125, 1, 0, 0, 0, 19, 128, 1, 0, 0, 0,
		21, 131, 1, 0, 0, 0, 23, 133, 1, 0, 0, 0, 25, 135, 1, 0, 0, 0, 27, 137,
		1, 0, 0, 0, 29, 140, 1, 0, 0, 0, 31, 143, 1, 0, 0, 0, 33, 146, 1, 0, 0,
		0, 35, 149, 1, 0, 0, 0, 37, 151, 1, 0, 0, 0, 39, 153, 1, 0, 0, 0, 41, 155,
		1, 0, 0, 0, 43, 157, 1, 0, 0, 0, 45, 159, 1, 0, 0, 0, 47, 161, 1, 0, 0,
		0, 49, 163, 1, 0, 0, 0, 51, 166, 1, 0, 0, 0, 53, 54, 5, 116, 0, 0, 54,
		55, 5, 114, 0, 0, 55, 56, 5, 117, 0, 0, 56, 63, 5, 101, 0, 0, 57, 58, 5,
		102, 0, 0, 58, 59, 5, 97, 0, 0, 59, 60, 5, 108, 0, 0, 60, 61, 5, 115, 0,
		0, 61, 63, 5, 101, 0, 0, 62, 53, 1, 0, 0, 0, 62, 57, 1, 0, 0, 0, 63, 2,
		1, 0, 0, 0, 64, 65, 5, 110, 0, 0, 65, 66, 5, 117, 0, 0, 66, 67, 5, 108,
		0, 0, 67, 72, 5, 108, 0, 0, 68, 69, 5, 110, 0, 0, 69, 70, 5, 105, 0, 0,
		70, 72, 5, 108, 0, 0, 71, 64, 1, 0, 0, 0, 71, 68, 1, 0, 0, 0, 72, 4, 1,
		0, 0, 0, 73, 74, 5, 105, 0, 0, 74, 75, 5, 110, 0, 0, 75, 6, 1, 0, 0, 0,
		76, 77, 5, 110, 0, 0, 77, 78, 5, 111, 0, 0, 78, 79, 5, 116, 0, 0, 79, 8,
		1, 0, 0, 0, 80, 82, 7, 0, 0, 0, 81, 80, 1, 0, 0, 0, 82, 83, 1, 0, 0, 0,
		83, 81, 1, 0, 0, 0, 83, 84, 1, 0, 0, 0, 84, 10, 1, 0, 0, 0, 85, 87, 7,
		0, 0, 0, 86, 85, 1, 0, 0, 0, 87, 88, 1, 0, 0, 0, 88, 86, 1, 0, 0, 0, 88,
		89, 1, 0, 0, 0, 89, 90, 1, 0, 0, 0, 90, 92, 5, 46, 0, 0, 91, 93, 7, 0,
		0, 0, 92, 91, 1, 0, 0, 0, 93, 94, 1, 0, 0, 0, 94, 92, 1, 0, 0, 0, 94, 95,
		1, 0, 0, 0, 95, 12, 1, 0, 0, 0, 96, 102, 5, 39, 0, 0, 97, 101, 8, 1, 0,
		0, 98, 99, 5, 92, 0, 0, 99, 101, 5, 39, 0, 0, 100, 97, 1, 0, 0, 0, 100,
		98, 1, 0, 0, 0, 101, 104, 1, 0, 0, 0, 102, 100, 1, 0, 0, 0, 102, 103, 1,
		0, 0, 0, 103, 105, 1, 0, 0, 0, 104, 102, 1, 0, 0, 0, 105, 117, 5, 39, 0,
		0, 106, 112, 5, 34, 0, 0, 107, 111, 8, 2, 0, 0, 108, 109, 5, 92, 0, 0,
		109, 111, 5, 34, 0, 0, 110, 107, 1, 0, 0, 0, 110, 108, 1, 0, 0, 0, 111,
		114, 1, 0, 0, 0, 112, 110, 1, 0, 0, 0, 112, 113, 1, 0, 0, 0, 113, 115,
		1, 0, 0, 0, 114, 112, 1, 0, 0, 0, 115, 117, 5, 34, 0, 0, 116, 96, 1, 0,
		0, 0, 116, 106, 1, 0, 0, 0, 117, 14, 1, 0, 0, 0, 118, 122, 7, 3, 0, 0,
		119, 121, 7, 4, 0, 0, 120, 119, 1, 0, 0, 0, 121, 124, 1, 0, 0, 0, 122,
		120, 1, 0, 0, 0, 122, 123, 1, 0, 0, 0, 123, 16, 1, 0, 0, 0, 124, 122, 1,
		0, 0, 0, 125, 126, 5, 38, 0, 0, 126, 127, 5, 38, 0, 0, 127, 18, 1, 0, 0,
		0, 128, 129, 5, 124, 0, 0, 129, 130, 5, 124, 0, 0, 130, 20, 1, 0, 0, 0,
		131, 132, 5, 33, 0, 0, 132, 22, 1, 0, 0, 0, 133, 134, 5, 62, 0, 0, 134,
		24, 1, 0, 0, 0, 135, 136, 5, 60, 0, 0, 136, 26, 1, 0, 0, 0, 137, 138, 5,
		62, 0, 0, 138, 139, 5, 61, 0, 0, 139, 28, 1, 0, 0, 0, 140, 141, 5, 60,
		0, 0, 141, 142, 5, 61, 0, 0, 142, 30, 1, 0, 0, 0, 143, 144, 5, 61, 0, 0,
		144, 145, 5, 61, 0, 0, 145, 32, 1, 0, 0, 0, 146, 147, 5, 33, 0, 0, 147,
		148, 5, 61, 0, 0, 148, 34, 1, 0, 0, 0, 149, 150, 5, 63, 0, 0, 150, 36,
		1, 0, 0, 0, 151, 152, 5, 58, 0, 0, 152, 38, 1, 0, 0, 0, 153, 154, 5, 40,
		0, 0, 154, 40, 1, 0, 0, 0, 155, 156, 5, 41, 0, 0, 156, 42, 1, 0, 0, 0,
		157, 158, 5, 91, 0, 0, 158, 44, 1, 0, 0, 0, 159, 160, 5, 93, 0, 0, 160,
		46, 1, 0, 0, 0, 161, 162, 5, 46, 0, 0, 162, 48, 1, 0, 0, 0, 163, 164, 5,
		44, 0, 0, 164, 50, 1, 0, 0, 0, 165, 167, 7, 5, 0, 0, 166, 165, 1, 0, 0,
		0, 167, 168, 1, 0, 0, 0, 168, 166, 1, 0, 0, 0, 168, 169, 1, 0, 0, 0, 169,
		170, 1, 0, 0, 0, 170, 171, 6, 25, 0, 0, 171, 52, 1, 0, 0, 0, 13, 0, 62,
		71, 83, 88, 94, 100, 102, 110, 112, 116, 122, 168, 1, 6, 0, 0,
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

// RiskDSLLexerInit initializes any static state used to implement RiskDSLLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewRiskDSLLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func RiskDSLLexerInit() {
	staticData := &RiskDSLLexerLexerStaticData
	staticData.once.Do(riskdsllexerLexerInit)
}

// NewRiskDSLLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewRiskDSLLexer(input antlr.CharStream) *RiskDSLLexer {
	RiskDSLLexerInit()
	l := new(RiskDSLLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &RiskDSLLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "RiskDSL.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// RiskDSLLexer tokens.
const (
	RiskDSLLexerBOOL_LIT  = 1
	RiskDSLLexerNULL_LIT  = 2
	RiskDSLLexerIN        = 3
	RiskDSLLexerNOT_KW    = 4
	RiskDSLLexerINT_LIT   = 5
	RiskDSLLexerFLOAT_LIT = 6
	RiskDSLLexerSTRING    = 7
	RiskDSLLexerID        = 8
	RiskDSLLexerAND       = 9
	RiskDSLLexerOR        = 10
	RiskDSLLexerNOT       = 11
	RiskDSLLexerGT        = 12
	RiskDSLLexerLT        = 13
	RiskDSLLexerGTE       = 14
	RiskDSLLexerLTE       = 15
	RiskDSLLexerEQ        = 16
	RiskDSLLexerNEQ       = 17
	RiskDSLLexerQMARK     = 18
	RiskDSLLexerCOLON     = 19
	RiskDSLLexerLPAREN    = 20
	RiskDSLLexerRPAREN    = 21
	RiskDSLLexerLBRACK    = 22
	RiskDSLLexerRBRACK    = 23
	RiskDSLLexerDOT       = 24
	RiskDSLLexerCOMMA     = 25
	RiskDSLLexerWS        = 26
)
