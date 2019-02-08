package aconf

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"text/scanner"
	"time"
	"unicode"
)

type HoconToken struct {
	Type  HoconTokenType
	Value string
	LexLocation
}

type HoconTokenType uint8

var forbiddenCharactersRegEx *regexp.Regexp

const NL = 0x0A
const HASH = 0x23
const HoconWS = 0x20

const (
	Integer HoconTokenType = iota
	Float
	Boolean
	Identifier
	Duration
	Size
	Key
	Text
	LeftBrace
	RightBrace
	LeftBracket
	RightBracket
	LeftParen
	RightParen
	Equals
	Colon
	Comma
	NewLine
	Other
)

type HoconLexer struct {
	scanner       scanner.Scanner
	previousToken HoconToken
	currentToken  HoconToken
	err           error
}

// NewLexer instantiates a new HoconLexer using the provided io.Reader
// Returns an error in case of
func NewLexer(reader io.Reader) (*HoconLexer, error) {
	var lexer HoconLexer
	var err error
	if reader == nil {
		return nil, &ErrReaderNil{}
	}
	lexer = HoconLexer{}
	lexer.scanner = scanner.Scanner{}
	fileContents, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	replacer := strings.NewReplacer(`"""`, "`")
	s := replacer.Replace(string(fileContents))
	r := strings.NewReader(s)
	lexer.scanner.Init(r)

	isIdentRune := func(ch rune, i int) bool {
		//return ch == '@' || ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0
		// If an identifier contains forbidden characters, then return false

		return (ch == '-' || ch == '_' || ch == '.' || unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0)
	}
	lexer.scanner.IsIdentRune = isIdentRune

	errorHandler := func(s *scanner.Scanner, msg string) {
		lexer.err = &LexScannerErr{msg, LexLocation{lexer.scanner.Line, lexer.scanner.Column}}
	}
	lexer.scanner.Error = errorHandler
	return &lexer, err
}

var tokenTypeMap map[rune]HoconTokenType

func init() {
	var err error
	tokenTypeMap = make(map[rune]HoconTokenType)

	tokenTypeMap['{'] = LeftBrace
	tokenTypeMap['}'] = RightBrace
	tokenTypeMap['['] = LeftBracket
	tokenTypeMap[']'] = RightBracket
	tokenTypeMap[':'] = Colon
	tokenTypeMap['='] = Equals
	tokenTypeMap[','] = Comma
	tokenTypeMap[scanner.Ident] = Identifier
	tokenTypeMap[scanner.Int] = Integer
	tokenTypeMap[scanner.Float] = Float
	tokenTypeMap[scanner.String] = Text
	tokenTypeMap[scanner.RawString] = Text

	forbiddenCharactersRegEx, err = regexp.Compile(`(\$|\"|\{|\}|\[|\]|:|=|,|\+|#|` + "`" + `|\^|\?|!|@|\*|&|\/\/)`)
	if err != nil {
		panic("Could not parse forbidden characters Regex")
	}

}

/*
Run1 returns a list of acceptable tokens for the parser to perform syntactical checking.
It is upto the parser to perform concatenation of strings, checking for sequence of tokens, ensuring no dangling tokens are present etc.

Keys are anything to the left of an = or :

Values are unquoted strings and quoted strings to the right of an = or :

A Key/Path expression must always start on a new line

Trailing spaces in Values should be trimmed, unless they are in a quoted string.
*/
func (lexer *HoconLexer) Run1() ([]HoconToken, error) {
	var tokens []HoconToken
	var err error

	for token := lexer.scanner.Scan(); token != scanner.EOF && lexer.err == nil; token = lexer.scanner.Scan() {
		switch token {

		case '#': // Processing a comment
			for r := lexer.scanner.Peek(); r != NL && r != scanner.EOF; {
				r = lexer.scanner.Next()
			}
		}
	}

	return tokens, err
}

func (lexer *HoconLexer) Run() ([]HoconToken, error) {
	var tokens []HoconToken
	//var err error
	if lexer == nil {
		return nil, &ErrLexerNotInitialized{}
	}
	durationRegEx, err := regexp.Compile(`^(\d+)\s*(s|second|seconds|ms|milli|millis|millisecond|milliseconds|ns|nano|nanos|nanosecond|nanoseconds|us|micro|micros|microsecond|microseconds|m|minute|minutes|h|hour|hours|d|day|days|w|week|weeks)$`)
	if err != nil {
		return nil, err
	}
	sizeRegEx, err := regexp.Compile(`^(\d+)\s*(B|b|byte|bytes|kb|kB|Kb|KB|kilobyte|kilobytes|mb|mB|Mb|MB|megabyte|megabytes|gb|Gb|GB|gB|gigabyte|gigabytes)`)
	if err != nil {
		return nil, err
	}

	forbiddenCharactersRegEx, err := regexp.Compile(`(\$|\"|\{|\}|\[|\]|:|=|,|\+|#|` + "`" + `|\^|\?|!|@|\*|&|\/\/)`)
	if err != nil {
		return nil, err
	}

	for token := lexer.scanner.Scan(); token != scanner.EOF && lexer.err == nil; token = lexer.scanner.Scan() {

		var tokenValue string
		var tokenType HoconTokenType

		if err != nil {
			return nil, err
		}

		switch token {

		case '#': // Processing a comment
			for r := lexer.scanner.Peek(); r != NL && r != scanner.EOF; {
				r = lexer.scanner.Next()
			}
			hoconToken := HoconToken{Type: NewLine, Value: "NewLine"}
			tokens = append(tokens, hoconToken)
			continue
		case '{', '}', '[', ']', '=', ':', ',', scanner.Ident, scanner.Float, scanner.Int:
			tokenValue = lexer.scanner.TokenText()
			// If an Identifier contains any of the forbidden characters, then error out
			if capGroups := forbiddenCharactersRegEx.FindAllStringSubmatch(tokenValue, -1); token == scanner.Ident && capGroups != nil {
				return nil, &LexInvalidTokenErr{tokenValue, LexLocation{lexer.scanner.Line, lexer.scanner.Column}}
			}
			// Ignore the ':' or '=' if it is going to be followed by an opening Brace or '{'
			if (token == ':' || token == '=') && lexer.scanner.Peek() == '{' {
				continue
			}
			tokenValue = lexer.scanner.TokenText()
			tokenType = tokenTypeMap[token]
			if token == scanner.Ident || token == scanner.Float || token == scanner.Int {
				numTokens := len(tokens)

				if numTokens > 0 && (tokens[numTokens-1].Type == Equals || tokens[numTokens-1].Type == Colon) {
					var buffer = bytes.NewBuffer([]byte(tokenValue))

					// Keep concatenating values till NL or HASH or One of the forbidden characters is encountered
					for r := lexer.scanner.Peek(); r != NL && r != HASH && r != scanner.EOF && forbiddenCharactersRegEx.FindAllStringSubmatch(string(r), -1) == nil; r = lexer.scanner.Peek() {
						r = lexer.scanner.Next()
						buffer.WriteString(string(r))
					}
					tokenValue = strings.TrimSpace(buffer.String())
					// In case of numbers, if the tail of the concatenated string does not match any units, then just strip out the tail and put it in the

					// Is it a boolean
					if strings.HasPrefix(tokenValue, "true") || strings.HasPrefix(tokenValue, "false") {
						tokenType = Boolean
					} else if capGroups := durationRegEx.FindAllStringSubmatch(tokenValue, -1); capGroups != nil {
						// check if the value starts with a number & ends in duration/size units
						v := capGroups[0][1]
						u := capGroups[0][2]
						var unitScale time.Duration
						switch u {
						case "s", "second", "seconds":
							unitScale = time.Second
						case "ms", "milli", "millis", "millisecond", "milliseconds":
							unitScale = time.Millisecond
						case "us", "micro", "micros", "microsecond", "microseconds":
							unitScale = time.Microsecond
						case "ns", "nano", "nanos", "nanosecond", "nanoseconds":
							unitScale = time.Nanosecond
						case "m", "minute", "minutes":
							unitScale = time.Minute
						case "h", "hour", "hours":
							unitScale = time.Hour
						case "d", "day", "days":
							unitScale = 24 * time.Hour
						case "w", "week", "weeks":
							unitScale = 24 * 7 * time.Hour
						}

						if d, err := strconv.ParseFloat(v, 64); err == nil {
							if d >= 0 {
								x := unitScale * time.Duration(d)
								tokenValue = fmt.Sprintf("%d", x)
							} else {
								return nil, &ErrLexerInvalidDuration{d}
							}
						}
						tokenType = Duration
					} else if capGroups := sizeRegEx.FindAllStringSubmatch(tokenValue, -1); capGroups != nil {
						v := capGroups[0][1]
						u := capGroups[0][2]
						var unitScale int
						switch u {
						case "B", "b", "byte", "bytes":
							unitScale = 1
						case "kb", "KB", "kB", "Kb", "kilobyte", "kilobytes":
							unitScale = 1024
						case "mb", "MB", "mB", "Mb", "megabyte", "megabytes":
							unitScale = 1024 * 1024
						case "gb", "GB", "gB", "Gb", "gigabyte", "gigabytes":
							unitScale = 1024 * 1024 * 1024
						}
						if s, err := strconv.Atoi(v); err == nil {
							if s >= 0 {
								x := unitScale * s
								tokenValue = fmt.Sprintf("%d", x)
							} else {
								return nil, &ErrLexerInvalidSize{s}
							}
						}
						tokenType = Size
					}
				} else if token == scanner.Ident {
					tokenType = Key
				}
			}
		case scanner.String, scanner.RawString:
			tokenType = tokenTypeMap[token]
			tokenValue, lexer.err = strconv.Unquote(lexer.scanner.TokenText())
		case scanner.Char:
			lexer.err = &ErrLexerInvalidToken{lexer.scanner.TokenText(), LexLocation{lexer.scanner.Line, lexer.scanner.Column}}
			continue
		default:
			lexer.err = &ErrLexerInvalidToken{lexer.scanner.TokenText(), LexLocation{lexer.scanner.Line, lexer.scanner.Column}}
			continue
		}

		hoconToken := HoconToken{tokenType, tokenValue, LexLocation{lexer.scanner.Line, lexer.scanner.Column}}
		tokens = append(tokens, hoconToken)

		if lexer.scanner.Peek() == NL {
			hoconToken := HoconToken{Type: NewLine, Value: "NewLine"}
			tokens = append(tokens, hoconToken)
		}

	}
	return tokens, lexer.err
}
