package aconf

import (
	"testing"
)

var balancedParenTests = []struct {
	fileName string
	err      error
}{
	{"test_data/hocon.good.conf", nil},
	{"test_data/hocon.unbalanced.paren.conf", &ParserUnbalancedParenthesesErr{}},
}

func TestBalancedParentheses(t *testing.T) {
	for _, test := range balancedParenTests {
		parser := &HoconParser{}
		if err := parser.Parse(test.fileName); !errorsAreEqual(err, test.err) {
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
