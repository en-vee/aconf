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
	s = 100 KB
	t = 10 seconds	
	name = axlrate imdg # comment at end of value
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

var testTokenize = []struct {
	fileContents string
	err          error
}{
	{validTokens, nil},
	{multilineStringTokens, nil},
	{unterminatedLiteralTokens, &LexScannerErr{"", LexLocation{1, 8}}},
	{unrecognizedTokens, &LexInvalidTokenErr{"", LexLocation{3, 3}}},
}

func TestVariousTokenizeTypes(t *testing.T) {
	for _, testcase := range testTokenize {
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
