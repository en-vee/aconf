package aconf

import (
	"fmt"
	"strings"
	"text/scanner"
)

/*
HoconTokenType represents valid HOCON tokens
*/
type HoconTokenType uint8

/*
All the valid HOCON token types
*/
const (
	LeafElement HoconTokenType = iota
	NodeElement
	LeftBrace
	RightBrace
	LeftParen
	RightParen
	LeftBracket
	RightBracket
	Integer
	Duration
	Decimal
	Equals
	Colon
	Comma
	Hash
	Eof
	Error
	Identifier
	QuotedString
	Other
)

const NL = 0x000A

type HoconToken struct {
	tokenType  HoconTokenType
	tokenValue string
}

////////////////////////////////////////////
// Stringer interface methods
////////////////////////////////////////////
func (token HoconToken) String() string {
	return fmt.Sprintf("tokenType=%d, tokenValue=%s", token.tokenType, token.tokenValue)
}

type HoconLexer struct {
	InputString   string
	previousToken HoconToken
	currentToken  HoconToken
	scan          scanner.Scanner
	//items         chan HoconToken
	tokens []HoconToken
	errs   []error
}

// State Functions return the next action. The state is the current token being processed.
// After detecting the valid next token, the current token is put in the channel for consumption.
type stateFn func(*HoconLexer) stateFn

var tokenNextStateFunctionMap map[rune]stateFn

func init() {
	tokenNextStateFunctionMap = make(map[rune]stateFn)
	tokenNextStateFunctionMap['#'] = lexHash
	tokenNextStateFunctionMap['{'] = lexText
	tokenNextStateFunctionMap['}'] = lexText
	tokenNextStateFunctionMap['['] = lexText
	tokenNextStateFunctionMap[']'] = lexText
	tokenNextStateFunctionMap['='] = lexText
	tokenNextStateFunctionMap[':'] = lexText
	tokenNextStateFunctionMap[scanner.Ident] = lexText
	tokenNextStateFunctionMap[scanner.String] = lexString
	tokenNextStateFunctionMap[scanner.Int] = lexText
	tokenNextStateFunctionMap[scanner.Float] = lexText
	tokenNextStateFunctionMap[scanner.EOF] = nil
}

func lexString(lexer *HoconLexer) stateFn {

	var returnFn stateFn = lexText
	hoconToken := HoconToken{}
	hoconToken.tokenValue = strings.TrimSuffix(strings.TrimPrefix(lexer.scan.TokenText(), `"`), `"`)
	hoconToken.tokenType = QuotedString
	lexer.tokens = append(lexer.tokens, hoconToken)
	return returnFn
}

func lexHash(lexer *HoconLexer) stateFn {
	var returnFn stateFn
	for lexer.scan.Peek() != NL {
		lexer.scan.Scan()
	}

	returnFn = lexText
	return returnFn
}

func lexText(lexer *HoconLexer) stateFn {
	var returnFn stateFn = lexText
	hoconToken := HoconToken{}
	r := lexer.scan.Scan()
	hoconToken.tokenValue = lexer.scan.TokenText()
	hoconToken.tokenType = Other
	switch r {
	case scanner.EOF:
		hoconToken.tokenType = Eof
	case '#':
		hoconToken.tokenType = Hash
	case '[':
		hoconToken.tokenType = LeftBracket
	case '{':
		hoconToken.tokenType = LeftBrace
	case ']':
		hoconToken.tokenType = RightBracket
	case '}':
		hoconToken.tokenType = RightBrace
	case '=':
		hoconToken.tokenType = Equals
	case ':':
		hoconToken.tokenType = Colon
	case ',':
		hoconToken.tokenType = Comma
	case scanner.Int:
		hoconToken.tokenType = Integer
	case scanner.Float:
		hoconToken.tokenType = Decimal
	case scanner.Ident:
		hoconToken.tokenType = Identifier
	case scanner.String:
		hoconToken.tokenType = QuotedString
	default:
		// Anything other than these is treated as an invalid HOCON Token
		//panic("Invalid Token " + hoconToken.tokenValue)
		lexer.errs = append(lexer.errs, &LexInvalidTokenErr{lexer.scan.TokenText(), LexLocation{lexer.scan.Line, lexer.scan.Column}})
	}

	if hoconToken.tokenType != Eof && hoconToken.tokenType != Other && hoconToken.tokenType != Hash && hoconToken.tokenType != QuotedString {
		//lexer.items <- hoconToken
		lexer.tokens = append(lexer.tokens, hoconToken)
	}

	if hoconToken.tokenType != Other {
		returnFn, _ = tokenNextStateFunctionMap[r]
	}

	return returnFn
}

func lexLeftBrace(lexer *HoconLexer) stateFn {
	lexer.currentToken.tokenType = LeftBrace
	lexer.currentToken.tokenValue = lexer.scan.TokenText()
	return nil
}

func lexRightBrace(lexer *HoconLexer) stateFn {
	lexer.currentToken.tokenType = RightBrace
	lexer.currentToken.tokenValue = lexer.scan.TokenText()
	return nil
}

func lexEquals(lexer *HoconLexer) stateFn {
	lexer.currentToken.tokenType = Equals
	lexer.currentToken.tokenValue = lexer.scan.TokenText()
	return nil
}

func lexLeftBracket(lexer *HoconLexer) stateFn {
	return nil
}

func lexRightBracket(lexer *HoconLexer) stateFn {
	return nil
}

func lexInt(lexer *HoconLexer) stateFn {
	var returnFn stateFn

	return returnFn
}

/*
func (lexer *HoconLexer) Run1() chan HoconToken {

	lexer.scan = scanner.Scanner{}
	lexer.scan.Init(strings.NewReader(lexer.InputString))
	//lexer.scan.Mode = scanner.GoTokens
	//lexer.scan.Mode |= scanner.ScanFloats
	lexer.items = make(chan HoconToken)

	go lexer.run()
	return lexer.items
}
*/

func (lexer *HoconLexer) Run() ([]HoconToken, []error) {

	lexer.scan = scanner.Scanner{}
	lexer.scan.Init(strings.NewReader(lexer.InputString))

	lexer.run()
	return lexer.tokens, lexer.errs
}

func (lexer *HoconLexer) run() {

	for state := lexText; state != nil; {
		state = state(lexer)
	}

}
