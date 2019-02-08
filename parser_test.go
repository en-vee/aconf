package aconf

import (
	"fmt"
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

	singlePathExp = `name = "axlrate-imdg"`
	twoPathExp    = `name = "axlrate"
	axlrate {
		imdg {
			name = "axlrate-imdg"
			member-count = 10
		}
	}`
	duplicateKeys = `a { b = 10 }
	a.b = 20`
)

var unbalancedParenTests = []struct {
	contents string
	err      error
}{

	{unbalancedParenContents, &ParserUnbalancedParenthesesErr{}},
}

var pathExpressionTests = []string{
	twoPathExp,
	duplicateKeys,
	singlePathExp,
}

func TestPathExpression(t *testing.T) {
	for _, testcontents := range pathExpressionTests {
		parser := &HoconParser{}
		reader := strings.NewReader(testcontents)
		var m map[string]interface{}
		var err error
		if m, err = parser.Parse(reader); err != nil {
			t.Errorf("Failed path expression test. Expected : nil. Got : %v", err)
		}

		for k, v := range m {
			fmt.Println("key = ", k, "\tvalue = ", v)
		}
	}
}

func TestUnBalancedParentheses(t *testing.T) {
	for _, test := range unbalancedParenTests {
		parser := &HoconParser{}
		reader := strings.NewReader(test.contents)
		if _, err := parser.Parse(reader); !errorsAreEqual(err, test.err) {
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
