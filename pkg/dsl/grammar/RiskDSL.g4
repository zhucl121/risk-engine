// RiskDSL.g4 — ANTLR4 grammar for RiskEngine rule condition expressions.
//
// Design constraints:
//   - Non-Turing-complete: no loops, no assignments, no side effects
//   - All expressions must terminate and return a bool
//   - Compatible with existing antonmedv/expr conditions used in payment_rules.yaml
//   - Extensible via FunctionRegistry (no grammar change needed to add new functions)

grammar RiskDSL;

// ─── Parser Rules ────────────────────────────────────────────────────────────

// Top-level rule: a condition is a single expression followed by EOF.
condition
    : expr EOF
    ;

// Expressions ordered by DESCENDING precedence (highest first).
// In ANTLR4 left-recursive rules, earlier alternatives bind TIGHTER (higher precedence).
//
// Precedence table (highest to lowest):
//   1. primary: FuncCall, MapIndex, FieldAccess, Paren, Literal, Ident
//   2. unary:   !
//   3. compare: > < >= <= == !=
//   4. logical: && ||
expr
    : '!' operand=expr                                                   # UnaryNot
    | left=expr op=('>' | '<' | '>=' | '<=' | '==' | '!=') right=expr  # BinaryCompare
    | left=expr op=('&&' | '||') right=expr                             # BinaryLogical
    | callee=ID '(' argList? ')'                                         # FuncCall
    | map=ID '[' key=STRING ']'                                          # MapIndex
    | obj=ID ('.' field=ID)+                                             # FieldAccess
    | '(' inner=expr ')'                                                 # Paren
    | lit=literal                                                        # LiteralExpr
    | name=ID                                                            # IdentExpr
    ;

argList
    : expr (',' expr)*
    ;

literal
    : INT_LIT    # IntLiteral
    | FLOAT_LIT  # FloatLiteral
    | STRING     # StringLiteral
    | BOOL_LIT   # BoolLiteral
    ;

// ─── Lexer Rules ──────────────────────────────────────────────────────────────

// Boolean literals must appear before ID so they are not tokenised as identifiers.
BOOL_LIT : 'true' | 'false' ;

// Numeric literals.
INT_LIT  : [0-9]+ ;
FLOAT_LIT: [0-9]+ '.' [0-9]+ ;

// String literals: single-quoted or double-quoted, no escape sequences needed
// for the current risk DSL use cases (feature keys and list names are ASCII).
STRING
    : '\'' (~'\'' | '\\\'')* '\''
    | '"'  (~'"'  | '\\"')*  '"'
    ;

// Identifiers: letters, digits, underscore. Must come after keyword tokens.
ID : [a-zA-Z_][a-zA-Z0-9_]* ;

// Operators and punctuation.
AND : '&&' ;
OR  : '||' ;
NOT : '!'  ;
GT  : '>'  ;
LT  : '<'  ;
GTE : '>=' ;
LTE : '<=' ;
EQ  : '==' ;
NEQ : '!=' ;

LPAREN : '(' ;
RPAREN : ')' ;
LBRACK : '[' ;
RBRACK : ']' ;
DOT    : '.' ;
COMMA  : ',' ;

// Whitespace is skipped.
WS : [ \t\r\n]+ -> skip ;
