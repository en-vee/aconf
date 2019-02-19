package aconf

import (
	"io"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"
)

type HoconParser struct {
	kvmap  map[string]interface{}
	tokens []HoconToken
	fldNum int
}

func (parser *HoconParser) Parse(hoconContentReader io.Reader, v interface{}) (map[string]interface{}, error) {

	var err error

	var m map[string]interface{}

	//lexer := HoconLexer{Reader: hoconContentReader}
	lexer, err := NewLexer(hoconContentReader)
	if parser.tokens, err = lexer.Run(); err != nil || parser.tokens == nil {
		return nil, err
	}

	if err = validateSyntax(parser.tokens); err != nil {
		return nil, err
	}

	if err = parser.unmarshal(v); err != nil {
		return nil, err
	}

	/*
		if m, err = buildMap(tokens); err != nil {
			return nil, err
		}
	*/

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

/*

 */
func (parser *HoconParser) unmarshal(v interface{}) error {
	var err error
	// Check if rv kind is pointer, if not, then error out
	rv := reflect.ValueOf(v)
	if rv.IsNil() || rv.Kind() != reflect.Ptr {
		return &ParserInvalidTargetErr{}
	}
	err = parser.decode(rv.Elem())
	return err
}

func (parser *HoconParser) decode(v reflect.Value) error {
	var err error

	// Start with the first token and determine it's type
	i := 0
	currToken := parser.tokens[i]
	switch currToken.Type {
	case Key:
		if parser.tokens[i+1].Type == Equals || parser.tokens[i+1].Type == Colon {
			// Check if key text/tokenValue matches the name of the corresponding Value
			//if v.Type().Field(0).Name == currToken.Value {
			// Advance to i+2th element in slice
			parser.tokens = parser.tokens[2:]
			if nv := v.FieldByName(currToken.Value); nv.IsValid() {
				// Short-Circuit and Recurse
				if err := parser.decode(nv); err != nil {
					return err
				}
			}
		} else if parser.tokens[i+1].Type == LeftBrace {
			parser.tokens = parser.tokens[2:]
			if v := v.FieldByName(currToken.Value); v.IsValid() && v.Kind() == reflect.Struct {
				// The root has changed. We now move into the next/nested struct
				if err := parser.decode(v); err != nil {
					return err
				}
			}
		}
	case Integer:
		val, err := strconv.ParseInt(currToken.Value, 10, 64)
		v.SetInt(val)
		return err
	case Float:
		val, err := strconv.ParseFloat(currToken.Value, 64)
		v.SetFloat(val)
		return err
	case Duration:
		val, err := time.ParseDuration(currToken.Value + "ns")
		v.SetInt(int64(val))
		return err
	case Equals, Colon:
		// ok
	case Text:
		v.SetString(currToken.Value)
		return err
	case Identifier, LeftBrace:
		// Start of a struct
	case LeftBracket:
		// Start of an array
	case RightBrace, RightBracket, NewLine:
		// ok
	default:
		err = &ParserInvalidTokenTypeErr{currToken}
	}
	if len(parser.tokens) > 1 {
		parser.tokens = parser.tokens[1:]
		parser.decode(v)
	}
	return err
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

}
