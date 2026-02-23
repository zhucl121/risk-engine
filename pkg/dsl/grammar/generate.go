// Package grammar holds the ANTLR4 grammar source and the go:generate directive
// that produces the lexer/parser/visitor Go code into ../parser/.
//
// Prerequisites (run once per developer machine):
//   brew install antlr  (or download antlr-4.13.x-complete.jar manually)
//
// Usage:
//   go generate ./pkg/dsl/grammar/
//
// The generated files are committed to the repository so that consumers do not
// need the ANTLR tool installed at build time.

package grammar

//go:generate java -jar ~/antlr-4.13.2-complete.jar -Dlanguage=Go -visitor -no-listener -package parser -o ../parser RiskDSL.g4
