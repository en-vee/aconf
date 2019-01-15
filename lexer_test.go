package aconf

import (
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
)

const fileContents = `
name = "axlrate-imdg"
//?
axlrate { # Main block
	name = "axlrate-imdg"
	# Another comment
	//  # This is an invalid character
	imdg {
		timeout = 10 seconds # number of seconds
		name = "axlrate-imdg"
	}
}	
`

var testTokenize = []struct {
	fileContents string
	err          error
}{
	{unterminatedLiteralTokens, &LexScannerErr{"", LexLocation{1, 8}}},
	{unrecognizedTokens, &LexInvalidTokenErr{"", LexLocation{3, 3}}},
	{validTokens, nil},
}

func TestVariousTokenizeTypes(t *testing.T) {
	for _, testcase := range testTokenize {
		l := HoconLexer{Reader: strings.NewReader(testcase.fileContents)}
		if _, err := l.Run(); testcase.err != nil {
			ok := false
			switch x := err.(type) {
			case *LexScannerErr:
				e := testcase.err.(*LexScannerErr)
				ok = (e.lineNumber == x.lineNumber && e.columnNumber == x.columnNumber)
			case *LexInvalidTokenErr:
				e := testcase.err.(*LexInvalidTokenErr)
				ok = (e.lineNumber == x.lineNumber && e.columnNumber == x.columnNumber)
			default:
				ok = true
			}
			if !ok {
				t.Errorf("wanted = %v, got = %v", testcase.err, err)
			}
		}

	}
}
