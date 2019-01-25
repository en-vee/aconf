package aconf

import (
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
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
	NewLine
	Other
)

var currentTokenHasError = false

const NL = 0x000A

type HoconToken struct {
	tokenType  HoconTokenType
	tokenValue string
	LexLocation
}

////////////////////////////////////////////
// Stringer interface method(s)
////////////////////////////////////////////
func (token HoconToken) String() string {
	return fmt.Sprintf("tokenType=%d, tokenValue=%s", token.tokenType, token.tokenValue)
}

type HoconLexer struct {
	Reader        io.Reader
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

var err error

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
	tokenNextStateFunctionMap[scanner.RawString] = lexString
	tokenNextStateFunctionMap[scanner.Int] = lexText
	tokenNextStateFunctionMap[scanner.Float] = lexText
	tokenNextStateFunctionMap[scanner.EOF] = nil
}

func lexString(lexer *HoconLexer) stateFn {
	var returnFn stateFn = lexText
	var err error

	hoconToken := HoconToken{}
	//hoconToken.tokenValue = strings.TrimSuffix(strings.TrimPrefix(lexer.scan.TokenText(), `"`), `"`)
	hoconToken.tokenValue, err = strconv.Unquote(lexer.scan.TokenText())
	if err != nil {
		return nil
	}
	hoconToken.tokenType = QuotedString
	hoconToken.lineNumber = lexer.scan.Line
	hoconToken.columnNumber = lexer.scan.Column
	if currentTokenHasError == false {
		lexer.tokens = append(lexer.tokens, hoconToken)
	}
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
	currentTokenHasError = false

	r := lexer.scan.Scan()
	//r := lexer.scan.Next()
	if currentTokenHasError == true {
		lexer.tokens = nil
		return nil
	}
	hoconToken := HoconToken{Other, lexer.scan.TokenText(), LexLocation{lexer.scan.Line, lexer.scan.Column}}
	//hoconToken.tokenValue = lexer.scan.TokenText()
	//hoconToken.tokenType = Other
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
	case scanner.String, scanner.RawString:
		hoconToken.tokenType = QuotedString
	default:
		currentTokenHasError = true
		returnFn = nil
		lexer.tokens = nil
		//lexer.errs = append(lexer.errs, &LexInvalidTokenErr{lexer.scan.TokenText(), LexLocation{lexer.scan.Line, lexer.scan.Column}})
		err = &LexInvalidTokenErr{lexer.scan.TokenText(), LexLocation{lexer.scan.Line, lexer.scan.Column}}
	}

	if currentTokenHasError == false && hoconToken.tokenType != Eof && hoconToken.tokenType != Other && hoconToken.tokenType != Hash && hoconToken.tokenType != QuotedString {
		//lexer.items <- hoconToken
		lexer.tokens = append(lexer.tokens, hoconToken)
	}

	if hoconToken.tokenType != Other {
		returnFn, _ = tokenNextStateFunctionMap[r]
	}

	return returnFn
}

func (lexer *HoconLexer) Run() ([]HoconToken, error) {

	lexer.scan = scanner.Scanner{}
	fileContents, err := ioutil.ReadAll(lexer.Reader)
	if err != nil {
		return nil, err
	}
	replacer := strings.NewReplacer(`"""`, "`")
	s := replacer.Replace(string(fileContents))
	lexer.Reader = strings.NewReader(s)

	lexer.scan.Init(lexer.Reader)

	/*
		lexIsIdentRune := func(ch rune, i int) bool {
			var rc bool
			if ch == '"' || unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0 {
				rc = true
			}
			return rc
		}

		lexer.scan.IsIdentRune = lexIsIdentRune
	*/

	errorHandler := func(s *scanner.Scanner, msg string) {
		currentTokenHasError = true
		//err := LexInvalidTokenErr{lexer.scan.TokenText(), LexLocation{lexer.scan.Line, lexer.scan.Column}}
		err = &LexScannerErr{msg, LexLocation{lexer.scan.Line, lexer.scan.Column}}
		//lexer.errs = append(lexer.errs, &err)

	}
	lexer.scan.Error = errorHandler
	lexer.run()
	return lexer.tokens, err
}

// The State Machine loop
func (lexer *HoconLexer) run() {

	for state := lexText; state != nil; {
		state = state(lexer)
	}
}
