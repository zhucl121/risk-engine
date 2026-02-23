// RiskDSL.g4 — ANTLR4 grammar for RiskEngine rule condition expressions.
//
// Design constraints:
//   - Non-Turing-complete: no loops, no assignments, no side effects
//   - All expressions must terminate and return a value (bool for top-level condition)
//   - Compatible with existing antonmedv/expr conditions used in payment_rules.yaml
//   - Extensible via FunctionRegistry (no grammar change needed to add new functions)
//
// v2 additions:
//   - Array literals:    [1, 'a', true]
//   - In / not-in:       value in [a, b, c]   |   value not in [a, b, c]
//   - Ternary operator:  cond ? thenExpr : elseExpr  (lowest precedence)
//
// Precedence table (highest to lowest):
//   1. primary:  FuncCall, MapIndex, FieldAccess, Paren, ArrayLit, Literal, Ident
//   2. unary:    !
//   3. compare:  > < >= <= == !=
//   4. in / not in
//   5. logical:  && ||
//   6. ternary:  ? :

grammar RiskDSL;

// ─── Parser Rules ────────────────────────────────────────────────────────────

// Top-level rule: a condition is a single expression followed by EOF.
condition
    : expr EOF
    ;

// Expressions ordered by DESCENDING precedence (highest first).
// In ANTLR4 left-recursive rules, earlier alternatives bind TIGHTER (higher precedence).
// Precedence table for expr (highest-binding alternative listed FIRST):
//   1. primary (FuncCall, MapIndex, FieldAccess, ArrayLit, Paren, Literal, Ident)
//   2. UnaryNot  (!)
//   3. BinaryCompare  (> < >= <= == !=)
//   4. In / NotIn
//   5. BinaryLogical  (&& ||)
//   6. Ternary  (? :)   — right-associative, lowest precedence
//
// In ANTLR4 left-recursive alternatives, the FIRST alternative has the HIGHEST
// operator precedence (tightest binding).  Put atomic/primary forms first and
// the weakest operators last.
expr
    // Unary negation — tighter than all binary ops.
    : '!' operand=expr                                            # UnaryNot

    // Comparison — left-associative.
    | left=expr op=('>' | '<' | '>=' | '<=' | '==' | '!=') right=expr  # BinaryCompare

    // Membership test.
    | left=expr 'not' 'in' right=expr                            # NotIn
    | left=expr 'in' right=expr                                  # In

    // Logical — left-associative, lower precedence than compare.
    | left=expr op=('&&' | '||') right=expr                      # BinaryLogical

    // Ternary — lowest precedence, right-associative.
    | cond=expr '?' thenExpr=expr ':' elseExpr=expr              # Ternary

    // Function call — must come before MapIndex and IdentExpr to avoid ambiguity.
    | callee=ID '(' argList? ')'                                  # FuncCall

    // Map subscript: features['key']
    | map=ID '[' key=STRING ']'                                   # MapIndex

    // Dotted field access: obj.field  |  geoIP(ip).country
    | obj=ID ('.' field=ID)+                                      # FieldAccess

    // Array literal: [expr, expr, ...]
    | '[' argList? ']'                                            # ArrayLiteral

    // Parenthesised sub-expression.
    | '(' inner=expr ')'                                          # Paren

    // Literal value.
    | lit=literal                                                 # LiteralExpr

    // Bare identifier (request fields: amount, userID, ip, etc.)
    | name=ID                                                     # IdentExpr
    ;

argList
    : expr (',' expr)*
    ;

literal
    : INT_LIT    # IntLiteral
    | FLOAT_LIT  # FloatLiteral
    | STRING     # StringLiteral
    | BOOL_LIT   # BoolLiteral
    | NULL_LIT   # NullLiteral
    ;

// ─── Lexer Rules ──────────────────────────────────────────────────────────────

// Keywords — must appear before ID so they are tokenised as keywords, not identifiers.
BOOL_LIT : 'true' | 'false' ;
NULL_LIT : 'null' | 'nil' ;

// 'in' and 'not' are contextual keywords; they appear as alternatives in the
// parser rule so ANTLR handles them correctly even when used as identifiers.
IN  : 'in' ;
NOT_KW : 'not' ;

// Numeric literals.
INT_LIT  : [0-9]+ ;
FLOAT_LIT: [0-9]+ '.' [0-9]+ ;

// String literals: single-quoted or double-quoted with basic escaping.
STRING
    : '\'' (~'\'' | '\\\'')* '\''
    | '"'  (~'"'  | '\\"')*  '"'
    ;

// Identifiers: letters, digits, underscore. Must come after keyword tokens.
ID : [a-zA-Z_][a-zA-Z0-9_]* ;

// Operators and punctuation.
AND    : '&&' ;
OR     : '||' ;
NOT    : '!'  ;
GT     : '>'  ;
LT     : '<'  ;
GTE    : '>=' ;
LTE    : '<=' ;
EQ     : '==' ;
NEQ    : '!=' ;
QMARK  : '?'  ;
COLON  : ':'  ;

LPAREN : '(' ;
RPAREN : ')' ;
LBRACK : '[' ;
RBRACK : ']' ;
DOT    : '.' ;
COMMA  : ',' ;

// Whitespace is skipped.
WS : [ \t\r\n]+ -> skip ;
