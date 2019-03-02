package aconf

import (
	"io"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"
)

type HoconParser struct {
	tokens            []HoconToken
	sliceBracketStack []int
	insideSlice       bool
	returnControl     bool
}

func (parser *HoconParser) Parse(hoconContentReader io.Reader, v interface{}) error {

	var err error

	//lexer := HoconLexer{Reader: hoconContentReader}
	lexer, err := NewLexer(hoconContentReader)
	if parser.tokens, err = lexer.Run(); err != nil || parser.tokens == nil {
		return err
	}

	if err = validateSyntax(parser.tokens); err != nil {
		return err
	}

	if err = parser.unmarshal(v); err != nil {
		return err
	}

	/*
		if m, err = buildMap(tokens); err != nil {
			return nil, err
		}
	*/

	return err
}

func validateSyntax(tokens []HoconToken) error {
	var err error
	// Check for Balanced Parentheses
	if err = checkBalancedParens(tokens); err != nil {
		return err
	}

	// Validate if closing braces are only preceded by NL or a Value
	for i, token := range tokens {
		if token.Type == RightBrace && !(tokens[i-1].Type == Text || tokens[i-1].Type == NewLine || tokens[i-1].Type == RightBracket) {
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
		return &ParserInvalidTargetErr{got: rv.Kind().String(), want: reflect.Ptr.String()}
	}
	err = parser.decode(rv.Elem(), rv.Elem())
	return err
}

func (parser *HoconParser) FieldByName(fieldName string, v reflect.Value) reflect.Value {
	var nv reflect.Value
	// First try to lookup the field directly in the Value
	if nv = v.FieldByName(fieldName); nv.IsValid() {
		return nv
	}
	// If not found then try to look it up based on tags
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		// Does fieldName match the tag value ?
		sTag := t.Field(i).Tag
		if sName, ok := sTag.Lookup("hocon"); ok {
			if sName == fieldName {
				// Return the Value.Field based on this sName
				nv = v.FieldByName(t.Field(i).Name)
				break
			}
		}
	}
	return nv
}

// function decode runs the main loop iterating over tokens and invoking handler functions
func (parser *HoconParser) decode(v reflect.Value, pv reflect.Value) error {
	var err error

	for {
		token := parser.tokens[0]
		//fmt.Println("tokenValue =", token.Value)
		//fmt.Println("v =", v.Type(), "pv =", pv.Type())

		switch token.Type {
		case Key:
			//v = parser.FieldByName(token.Value, pv)
			v = parser.FieldByName(token.Value, pv)
		case Equals, Colon, Comma:
		case NewLine:
		case Boolean, Integer, Duration, Float, Text:
			if err = parser.setValue(v, token); err != nil {
				return err
			}
		case LeftBrace:
			/*
				Use root/parent value pv = v
				And then invoke decode object
			*/
			parser.tokens = parser.tokens[1:]
			if err = parser.decodeObject(v, v); err != nil {
				return err
			}
		case LeftBracket:
			// start of an slice
			// create the slice
			nv := reflect.MakeSlice(v.Type(), 0, 0)
			v.Set(nv)
			// Call decodeSequence
			parser.tokens = parser.tokens[1:]
			if err = parser.decodeSequence(v, v); err != nil {
				return err
			}
		case RightBrace:
		case RightBracket:
		}

		// Advance the tokens
		numTokens := len(parser.tokens)
		if numTokens <= 1 {
			break
		}
		parser.tokens = parser.tokens[1:]

	}
	return err
}

// function decodeObject is invoked when the caller sees a '{' and recurses till a matching '}' is found.
//
// Assumes that the '{' has already been parsed
func (parser *HoconParser) decodeObject(v reflect.Value, pv reflect.Value) error {
	var err error

	lCount := 1
	rCount := 0
	for {
		token := parser.tokens[0]
		//fmt.Println("v =", v.Type(), "pv =", pv.Type())
		switch token.Type {
		case Key:
			v = parser.FieldByName(token.Value, pv)
		case Equals, Colon, Comma:
		case NewLine:
		case Boolean, Integer, Duration, Float, Text:
			if err = parser.setValue(v, token); err != nil {
				return err
			}
		case LeftBrace:
			lCount++
			parser.tokens = parser.tokens[1:] //this is to avoid the infinite loop
			if err = parser.decodeObject(v, v); err != nil {
				return err
			}
			// We have returned from decodeObject, so update the rCount otherwise we never return
			rCount++
		case LeftBracket:
			// start of an slice
			// create the slice
			nv := reflect.MakeSlice(v.Type(), 0, 0)
			v.Set(nv)
			// Call decodeSequence
			parser.tokens = parser.tokens[1:] //this is to avoid the infinite loop
			if err = parser.decodeSequence(v, v); err != nil {
				return err
			}
		case RightBrace:
			rCount++
		case RightBracket:
		}
		// Advance the tokens
		numTokens := len(parser.tokens)
		if numTokens <= 1 {
			break
		}

		if lCount == rCount {
			// Reached end of block
			break
		}
		parser.tokens = parser.tokens[1:]

	}
	return err
}

// function decodeSequence handles aggregate types arrays/slices and is called by the decode function.
// It can call the function decodeObject, if the aggregate type is an object
// assumes that '[' has already been scanned and the first character is the one after that
func (parser *HoconParser) decodeSequence(v reflect.Value, pv reflect.Value) error {
	var err error

	lCount := 1
	rCount := 0
	index := 0
	for {
		token := parser.tokens[0]
		//fmt.Println("v =", v.Type(), "pv =", pv.Type())
		switch token.Type {
		case Key:
			v = parser.FieldByName(token.Value, pv)
		case Equals, Colon, Comma, NewLine:
			//Ok
		case Boolean, Integer, Duration, Float, Text:
			// Allocate a new element
			elem := reflect.New(pv.Type().Elem()).Elem()
			// Append the element

			pv.Set(reflect.Append(pv, elem))
			v = pv.Index(pv.Len() - 1)
			//v = pv.Index(index)
			if err = parser.setValue(v, token); err != nil {
				return err
			}
			index++
		case LeftBrace:
			// Allocate a new element
			elem := reflect.New(pv.Type().Elem()).Elem()
			pv.Set(reflect.Append(pv, elem))
			v = pv.Index(pv.Len() - 1)
			parser.tokens = parser.tokens[1:]
			if err = parser.decodeObject(v, v); err != nil {
				return err
			}

		case LeftBracket:
			lCount++
			// start of an slice
			// create the slice
			nv := reflect.MakeSlice(v.Type(), 0, 0)
			v.Set(nv)
			// Recursive call
			if err = parser.decodeSequence(v, v); err != nil {
				return err
			}
			rCount++
		case RightBrace:

		case RightBracket:
			rCount++
		}
		// Advance the tokens
		numTokens := len(parser.tokens)
		if numTokens <= 1 {
			break
		}

		if lCount == rCount {
			// Reached end of block
			break
		}
		parser.tokens = parser.tokens[1:]
	}
	return err
}

func (parser *HoconParser) getNextToken(startIndex int) (HoconToken, int) {
	var token HoconToken
	var index int
	for i, t := range parser.tokens[startIndex:] {
		if t.Type != NewLine {
			token = t
			index = i
			break
		}
	}
	return token, index
}

func (parser *HoconParser) setValue(v reflect.Value, token HoconToken) error {
	var err error

	tokenValue := token.Value
	switch token.Type {
	case Boolean:
		val, err := strconv.ParseBool(tokenValue)
		if err != nil {
			return err
		}
		if v.IsValid() && v.Kind() == reflect.Bool {
			v.SetBool(val)
		}
	case Integer:
		val, err := strconv.ParseInt(tokenValue, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(val)
	case Float:
		val, err := strconv.ParseFloat(tokenValue, 64)
		if err != nil {
			return err
		}
		v.SetFloat(val)
	case Duration:
		val, err := time.ParseDuration(tokenValue + "ns")
		if err != nil {
			return err
		}
		v.SetInt(int64(val))
	case Text:
		if v.IsValid() && v.Kind() == reflect.String {
			v.SetString(tokenValue)
		}
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
