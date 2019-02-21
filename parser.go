package aconf

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type HoconParser struct {
	kvmap        map[string]interface{}
	tokens       []HoconToken
	oldRootValue reflect.Value
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

	// Array data type validation
	// Determine Type of array based on first element in it
	// If all are not same, then error out

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

func (parser *HoconParser) FieldByName(fieldName string, v reflect.Value) reflect.Value {
	var nv reflect.Value
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Type().Field(i).Name == strings.ToTitle(fieldName) {
			nv = v.Field(i)
			fmt.Println(v)
			fmt.Println(nv)
			fmt.Println(v.FieldByName(fieldName))
			break
		}
	}
	return nv
}

func (parser *HoconParser) decode(v reflect.Value) error {
	var err error

	// Start with the first token and determine it's type
	i := 0
	currToken := parser.tokens[i]
	switch currToken.Type {
	case Key:
		if (parser.tokens[i+1].Type == Equals || parser.tokens[i+1].Type == Colon) && parser.tokens[i+2].Type != LeftBracket {
			// Check if key text/tokenValue matches the name of the corresponding Value
			//if v.Type().Field(0).Name == currToken.Value {
			// Advance to i+2th element in slice
			parser.tokens = parser.tokens[2:]

			//if nv := v.FieldByName(currToken.Value); nv.IsValid() {
			if nv := v.FieldByName(currToken.Value); nv.IsValid() {
				//if nv := parser.FieldByName(currToken.Value, v); nv.IsValid() {
				// Short-Circuit and Recurse
				if err := parser.decode(nv); err != nil {
					return err
				}
			}
		} else if parser.tokens[i+1].Type == LeftBrace {
			// Move to the first token which is non-NewLine
			for index, t := range parser.tokens {
				if t.Type != NewLine {
					parser.tokens = parser.tokens[index:]
					break
				}
			}
			// Check if struct field name matches the current token value
			if nv := v.FieldByName(currToken.Value); nv.IsValid() && nv.Kind() == reflect.Struct {
				// The root has changed. We now descend into the next/nested struct
				parser.oldRootValue = v
				if err := parser.decode(nv); err != nil {
					return err
				}
			}
		} else if (parser.tokens[i+1].Type == Equals || parser.tokens[i+1].Type == Colon) && parser.tokens[i+2].Type == LeftBracket {
			// (parser.tokens[i+1].Type == Equals || parser.tokens[i+1].Type == Colon) && parser.tokens[i+2].Type == LeftBracket
			// This is a key which points to an array
			if nv := v.FieldByName(currToken.Value); nv.IsValid() {
				parser.tokens = parser.tokens[i+3:]
				if err := parser.decodeSequence(nv); err != nil {
					return err
				}
			}
		}
	case Integer:
		val, err := strconv.ParseInt(currToken.Value, 10, 64)
		if err != nil {
			return err
		}
		//fmt.Println(v.Type())
		v.SetInt(val)
		return err
	case Float:
		val, err := strconv.ParseFloat(currToken.Value, 64)
		if err != nil {
			return err
		}
		//fmt.Println(v.Type())
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
	case NewLine:
		// OK
	case RightBrace, RightBracket:
		// ok
		// Rewind to the old root
		v = parser.oldRootValue
	default:
		err = &ParserInvalidTokenTypeErr{currToken}
	}
	if len(parser.tokens) > 1 {
		parser.tokens = parser.tokens[1:]
		parser.decode(v)
	}
	return err
}

// Assumes that '[' has already been scanned and first token is the one after '['
func (parser *HoconParser) decodeSequence(v reflect.Value) error {
	var err error
	currentToken := parser.tokens[0]
	//prevToken := currentToken
	if currentToken.Type == RightBracket {
		// End of Array
	} else if currentToken.Type == LeftBrace {
		// Array of Objects
	} else {
		// Array of primitives
		// X = 10
		// Y = [ 1, 2, 3, 4 ]
		// Y []int
		// v should be pointing to Y, so depending on data type of currentToken, keep appending to the slice till an invalid token / end of array is encountered
		if v.Kind() != reflect.Slice {
			// TODO : Return Error
		}

		// Count the number of array elements
		numElements := 0
		arrayEndIndex := 0
		for i, token := range parser.tokens {
			if token.Type == RightBracket {
				arrayEndIndex = i
				break
			}
			if token.Type != Comma {
				numElements++
			}
		}

		if nv := reflect.MakeSlice(v.Type(), numElements, numElements); nv.IsValid() {

			i := 0

			for _, token := range parser.tokens {
				if token.Type == RightBracket {
					parser.tokens = parser.tokens[arrayEndIndex:]
					break
				}
				if token.Type != Comma {
					// Decode the token value into the current slice element
					switch token.Type {
					case Integer:
						val, err := strconv.ParseInt(token.Value, 10, 64)
						if err != nil {
							return err
						}
						//fmt.Println(v.Type())
						nv.Index(i).SetInt(val)
						i++
					}
				}
			}

			v.Set(nv)
		}

		// Call decode method to recurse
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
