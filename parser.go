package aconf

import (
	"io"
	"strconv"
	"unicode/utf8"
)

var m map[string]interface{}

type HoconParser struct {
	m map[string]interface{}
}

func (parser *HoconParser) Parse(hoconContentReader io.Reader) (map[string]interface{}, error) {

	var err error
	var tokens []HoconToken
	var m map[string]interface{}

	//lexer := HoconLexer{Reader: hoconContentReader}
	lexer, err := NewLexer(hoconContentReader)
	if tokens, err = lexer.Run(); err != nil {
		return nil, err
	}

	if err = validateSyntax(tokens); err != nil {
		return nil, err
	}

	if m, err = buildMap(tokens); err != nil {
		return nil, err
	}

	return m, err
}

func validateSyntax(tokens []HoconToken) error {
	var err error
	// Check for Balanced Parentheses
	if err = checkBalancedParens(tokens); err != nil {
		return err
	}

	// Validate if closing braces are only preceded by NL or a Value
	for i, token := range tokens {
		if token.Type == RightBrace && !(tokens[i-1].Type == Text || tokens[i-1].Type == NewLine) {
			err = &LexInvalidTokenErr{tokens[i-1].Value, tokens[i-1].LexLocation}
			break
		}
	}

	// Validate if opening square-brackets are preceded by an equals or colon token
	for i, token := range tokens {
		if token.Type == LeftBracket && !(tokens[i-1].Type == Equals || tokens[i-1].Type == Colon) {
			err = &ParserInvalidArrayErr{tokens[i-1].Value, tokens[i-1].LexLocation}
			break
		}
	}

	// Validate there are no dangling keys i.e. keys followed by nothing

	return err
}

// Populate a map of the paths/keys (string) -> values (interface{})
func buildMap(tokens []HoconToken) (map[string]interface{}, error) {
	var err error
	var keyPath, prevKeyPath string
	var m = make(map[string]interface{})

	//keyPath = tokens[0].Value
	for i, t := range tokens {
		if t.Type == Key {
			prevKeyPath = keyPath
			if keyPath == "" {
				keyPath = t.Value
			} else {
				keyPath = keyPath + "." + t.Value
			}
		}
		if (len(tokens)-(i+1)) >= 2 && tokens[i+1].Type == Equals {
			val := tokens[i+2]
			switch val.Type {
			case Integer:
				m[keyPath], err = strconv.ParseInt(val.Value, 0, 64)
			case Float, Duration, Size:
				m[keyPath], err = strconv.ParseFloat(val.Value, 64)
			default:
				m[keyPath] = val.Value
			}

			keyPath = prevKeyPath
		}

		if t.Type == RightBrace {
			keyPath = ""
		}
	}

	return m, err
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
}
