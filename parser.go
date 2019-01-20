package aconf

import (
	"os"
	"unicode/utf8"
)

var m map[string]interface{}

type HoconParser struct {
}

func (parser *HoconParser) Parse(fileName string) error {

	var err error
	// Check if file has any / characters
	// If there are, then validate if the file exists. If it does not, then exit with error
	// If no / characters, then validate if file exists in present working directory
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	lexer := HoconLexer{Reader: f}
	tokens, err := lexer.Run()
	if err = checkBalancedParens(tokens); err != nil {
		return err
	}
	return err
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
