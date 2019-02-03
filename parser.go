package aconf

import (
	"io"
	"unicode/utf8"
)

var m map[string]interface{}

type HoconParser struct {
}

func (parser *HoconParser) Parse(hoconContentReader io.Reader) error {

	var err error
	var tokens []HoconToken

	//lexer := HoconLexer{Reader: hoconContentReader}
	lexer, err := NewLexer(hoconContentReader)
	if tokens, err = lexer.Run(); err != nil {
		return err
	}

	if err = checkBalancedParens(tokens); err != nil {
		return err
	}

	if err = buildMap(tokens); err != nil {
		return err
	}

	return err
}

// Populate a map of the keys (string) -> values (interface{})
func buildMap(tokens []HoconToken) error {
	var err error
	var keyPath string
	numTokens := len(tokens)
	keyPath = tokens[0].Value
	for index, token := range tokens {

		/*
			Determine if current token is a leaf key and then add it to the map, if not, then keep appending to the path
			- If it is followed by an =,:
			- If it is followed by a { which is then followed by an identifier and an equals
		*/
		cond1 := (token.Type == Identifier)
		cond2 := ((index + 1) < numTokens) && (tokens[index+1].Type == Equals || tokens[index+1].Type == Colon)
		cond3 := ((index + 2) < numTokens) && (tokens[index+2].Type == LeftBrace || token.Type == LeftBracket)
		cond4 := ((index + 2) < numTokens) && (tokens[index+2].Type == Number || tokens[index+2].Type == Text)
		if cond1 {
			if cond2 && (cond3 || cond4) {

				m[keyPath] = tokens[index+2].Value
				// Path building can be terminated
				keyPath = ""

			} else {
				// Keep building the path
				keyPath = keyPath + "." + token.Value
			}
		}
	}
	return err
}

func buildPath() {

}

func checkBalancedParens(tokens []HoconToken) error {
	var err error
	var stack []rune
	for _, token := range tokens {
		switch token.Type {
		case LeftBrace, LeftBracket, LeftParen:
			// Push it on the stack
			r, size := utf8.DecodeRuneInString(token.Value)
			if r == utf8.RuneError && (size == 0 || size == 1) {
				// Return error
				return &ParserInvalidRuneErr{}
			}
			stack = append(stack, r)
		case RightBrace, RightBracket, RightParen:
			// Pop it from the stack
			l := len(stack)
			stack = stack[:l-1]
		}
	}
	if len(stack) != 0 {
		err = &ParserUnbalancedParenthesesErr{}
	}
	return err
}

func init() {
	m = make(map[string]interface{})
	m[""] = 10
	m[""] = "abc"
}
