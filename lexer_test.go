package aconf

import (
	"fmt"
	"strings"
	"testing"
)

const (
	unterminatedLiteralTokens = `name = "axlrate-`
	unrecognizedTokens        = `
	{
		*
		name = "axlrate"
	}
	`
	validTokens = `
name = "axlrate-imdg"
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
	line2
	"""
	`
)

var testTokenize = []struct {
	fileContents string
	err          error
}{
	{multilineStringTokens, nil},
	{unterminatedLiteralTokens, &LexScannerErr{"", LexLocation{1, 8}}},
	{unrecognizedTokens, &LexInvalidTokenErr{"", LexLocation{3, 3}}},
	{validTokens, nil},
}

func TestVariousTokenizeTypes(t *testing.T) {
	for _, testcase := range testTokenize {
		reader := strings.NewReader(testcase.fileContents)
		l := HoconLexer{Reader: reader}
		if tokens, err := l.Run(); err != nil {
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
				fmt.Printf("%v\n", token)
			}
		}

	}
}
