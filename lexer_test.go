package aconf

import (
	"strings"
	"testing"
)

const (
	tokensWithDurationUnits = `
	timeOut = 10 seconds
	`
	tokenWithSizeUnits                = `size = 5 GB`
	tokensWithUnquotedValues          = `name = axlrate imdg`
	tokensWithCommentsOnSeparateLines = `
	# First Line
	x = 10`
	tokensWithCommentsAtEndOfValue = `name = axlrate-imdg # This is the grid name`
	tokenWithHyphenatedKey         = `grid-name = axlrate-imdg`
	unterminatedLiteralTokens      = `name = "axlrate-`
	unrecognizedTokens             = `
	{
		*
		name = "axlrate"
	}
	`
	validTokens = `
	s = 100 KB
	t = 10 seconds	
	name = axlrate imdg # comment at end of un-quoted value
	key = "quoted string value"
	# First line comment
//?
axlrate { # Main block
	name = "axlrate-imdg"
	# Another comment
	// * # This is an invalid character
	imdg {
		timeout = 10 seconds # number of seconds
		name = "axlrate-imdg"
	}
}
	`
	multilineStringTokens = `x = """
line1
"quoted-and-embedded-line"
line2
"""
	`
)

var testInvalidTokens = []struct {
	fileContents string
	err          error
}{
	{unterminatedLiteralTokens, &LexScannerErr{"", LexLocation{1, 8}}},
	{unrecognizedTokens, &LexInvalidTokenErr{"", LexLocation{3, 3}}},
}

var testValidTokens = []struct {
	fileContents string
	tokenIndices []int
	key          string
	value        string
	valueType    HoconTokenType
}{
	{tokensWithCommentsOnSeparateLines, []int{1, 3}, "x", "10", Number},
	{tokenWithHyphenatedKey, []int{0, 2}, "grid-name", "axlrate-imdg", Identifier},
	{tokensWithDurationUnits, []int{0, 2}, "timeOut", "10000000000", Duration},
	{tokenWithSizeUnits, []int{0, 2}, "size", "5368709120", Size},
	{tokensWithCommentsAtEndOfValue, []int{0, 2}, "name", "axlrate-imdg", Identifier},
	{tokensWithUnquotedValues, []int{0, 2}, "name", "axlrate imdg", Identifier},
	{multilineStringTokens, []int{0, 2}, "x", "\nline1\n\"quoted-and-embedded-line\"\nline2\n", Text},
}

func TestValidTokens(t *testing.T) {

	for _, testcase := range testValidTokens {
		reader := strings.NewReader(testcase.fileContents)
		l, _ := NewLexer(reader)
		var tokens []HoconToken
		var err error
		if tokens, err = l.Run(); err != nil {
			t.Errorf("test failed with non-nil error : Expected : Nil, Got : %v", err)
		}

		if !(tokens[testcase.tokenIndices[0]].Value == testcase.key) {
			t.Errorf("Mismatched Keys -> Got: %v, Want: %v", testcase.key, tokens[testcase.tokenIndices[0]].Value)
		}

		if !(tokens[testcase.tokenIndices[1]].Value == testcase.value) {
			t.Errorf("Mismatched Values -> Got: %v, Want: %v", testcase.value, tokens[testcase.tokenIndices[1]].Value)
		}

		if !(tokens[testcase.tokenIndices[1]].Type == testcase.valueType) {
			t.Errorf("Mismatched Types -> Got: %v, Want: %v", testcase.valueType, tokens[testcase.tokenIndices[1]].Type)
		}
	}
}

func TestInvalidTokens(t *testing.T) {
	for _, testcase := range testInvalidTokens {
		reader := strings.NewReader(testcase.fileContents)
		//l := HoconLexer{Reader: reader}
		l, _ := NewLexer(reader)
		if tokens, err := l.Run(); err != nil {
			t.Logf("Error : %v\n", err)
			ok1 := false
			// Brilliant example of type switches and assertions used in combination
			switch x := err.(type) {
			case *LexScannerErr:
				e, ok2 := testcase.err.(*LexScannerErr)
				ok1 = ok2 && (e.lineNumber == x.lineNumber && e.columnNumber == x.columnNumber)
			case *LexInvalidTokenErr:
				e, ok2 := testcase.err.(*LexInvalidTokenErr)
				ok1 = ok2 && (e.lineNumber == x.lineNumber && e.columnNumber == x.columnNumber)
			default:
				ok1 = true
			}
			if !ok1 {
				t.Errorf("wanted = %v, got = %v", testcase.err, err)
			}
		} else {
			for _, token := range tokens {
				t.Logf("%v\n", token)
			}
		}

	}
}
