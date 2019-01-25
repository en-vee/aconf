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

	lexer := HoconLexer{Reader: hoconContentReader}
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
	for index, token := range tokens {
		/*
			Determine if current token is a leaf key and then add it to the map, if not, then keep appending to the path
			- If it is followed by an =,:
			- If it is followed by a { which is then followed by an identifier and an equals
		*/
		if token.tokenType == Identifier && (tokens[index+1].tokenType == Equals || tokens[index+1].tokenType == Colon) {
			if tokens[index+2].tokenType != LeftBrace && tokens[index+2].tokenType != LeftParen && tokens[index+2].tokenType != LeftBracket {
				m[keyPath] = tokens[index+2].tokenValue
				// Path building can be terminated
				keyPath = ""
			}
		} else {
			// Keep building the path
			keyPath = keyPath + "." + token.tokenValue
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
		switch token.tokenType {
		case LeftBrace, LeftBracket, LeftParen:
			// Push it on the stack
			r, size := utf8.DecodeRuneInString(token.tokenValue)
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
