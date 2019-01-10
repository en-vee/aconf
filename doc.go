/*
Package hocon parses a HOCON file, validates it's syntax and creates a map of property keys and their values.
The lexer emits tokens on a channel for clients (parsers) of the lexer to consume.
The tokens emitted by the lexer consist of a struct :
type HoconToken struct {
	tokenType HoconTokenType
	tokenText string
}

The parser then uses these tokens and performs syntactical analysis.
For each token, it does the following
*/

package aconf
