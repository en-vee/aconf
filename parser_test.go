package aconf

import (
	"strings"
	"testing"
)

const (
	unbalancedParenContents = `
	name = "axlrate-imdg"
	axlrate { # Main block
	name = "axlrate-imdg"
	# Another comment
	//  # This is an invalid character
	imdg {
		timeout = 10 seconds # number of seconds
		name = "axlrate-imdg"
	}
//}`

	balancedParenContents = `
	name = "axlrate-imdg"
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
)

var balancedParenTests = []struct {
	contents string
	err      error
}{
	{balancedParenContents, nil},
	{unbalancedParenContents, &ParserUnbalancedParenthesesErr{}},
}

func TestBalancedParentheses(t *testing.T) {
	for _, test := range balancedParenTests {
		parser := &HoconParser{}
		reader := strings.NewReader(test.contents)
		if err := parser.Parse(reader); !errorsAreEqual(err, test.err) {
			t.Errorf("Expected : %v, Got : %v", test.err, err)
		}
	}
}

func errorsAreEqual(actual, expected error) bool {
	ok := false
	if actual == nil && expected == nil {
		ok = true
	}
	switch actual := actual.(type) {
	case *ParserUnbalancedParenthesesErr:
		e, matched := expected.(*ParserUnbalancedParenthesesErr)
		ok = matched && (actual.columnNumber == e.columnNumber) && actual.lineNumber == e.lineNumber
	case nil:
		ok = true
	}
	return ok
}
