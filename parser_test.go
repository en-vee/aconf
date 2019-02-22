package aconf

import (
	"strings"
	"testing"
	"time"
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

type TestTableStruct struct {
	contents     string
	target       interface{}
	validateFunc func(t interface{}) bool
}

type IntStruct struct {
	X int
}

type FloatStruct struct {
	X float64
}

type StringStruct struct {
	X string
}

type DurationStruct struct {
	X time.Duration
}

var singleKeyValuePairTests = []TestTableStruct{
	{contents: `X = 10`, target: &IntStruct{}, validateFunc: func(target interface{}) bool { return target.(*IntStruct).X == 10 }},
	{contents: `X = 10.4567`, target: &FloatStruct{}, validateFunc: func(target interface{}) bool { return target.(*FloatStruct).X == 10.4567 }},
	{contents: `X = 10 seconds`, target: &DurationStruct{}, validateFunc: func(target interface{}) bool { return target.(*DurationStruct).X == 10*time.Second }},
	{contents: `X = unquoted string`, target: &StringStruct{}, validateFunc: func(target interface{}) bool { return target.(*StringStruct).X == "unquoted string" }},
}

// TestSingleKeyValuePair tests parsing of simple assignments like x = 10 or a = "bcd"
func TestSingleKeyValuePair(t *testing.T) {
	for _, testcase := range singleKeyValuePairTests {
		//t.Log("Before : ", testcase.target)
		//t.Logf("input: %v", testcase.contents)
		parser := &HoconParser{}
		reader := strings.NewReader(testcase.contents)
		if err := parser.Parse(reader, testcase.target); err != nil {
			t.Errorf("failed for input : %v. Error : %v", testcase.contents, err)
		}
		if !testcase.validateFunc(testcase.target) {
			t.Errorf("input: %v", testcase.contents)
		}
		//t.Log("After : ", testcase.target)
	}
}

type MultiValueStruct struct {
	Y float64
	Z string
	X int
}

var multipleKeyValurPairsTests = []TestTableStruct{
	{contents: `X = 10
	Y = 10.897
	Z = unq uoted`, target: &MultiValueStruct{}, validateFunc: func(target interface{}) bool {
		v, ok := target.(*MultiValueStruct)
		return ok && v.X == 10 && v.Y == 10.897 && v.Z == "unq uoted"
	}},
}

func TestMultipleKeyValuePairs(t *testing.T) {
	for _, testcase := range multipleKeyValurPairsTests {
		//t.Log("Before : ", testcase.target)
		//t.Logf("input: %v", testcase.contents)
		parser := &HoconParser{}
		reader := strings.NewReader(testcase.contents)
		if err := parser.Parse(reader, testcase.target); err != nil {
			t.Errorf("failed for input : %v. Error : %v", testcase.contents, err)
		}
		if !testcase.validateFunc(testcase.target) {
			t.Errorf("input: %v", testcase.contents)
		}
		//t.Log("After : ", testcase.target)
	}
}

type structWithSingleInnerBlock struct {
	IntStruct
	FloatValues struct {
		FloatStruct
	}
	StringValues struct {
		StringStruct
	}
	DurationValues struct {
		DurationStruct
	}
}

type SysConfig struct {
	Cluster struct {
		Name       string
		Connection struct {
			Host    string
			Port    int
			Timeout time.Duration
		}
	}
}

type IntArrayStruct struct {
	X []int
}

type ArrayInsideStruct struct {
	A string
	B struct {
		X []int
	}
}

type ArrayInsideStructFollowedByKeyValPair struct {
	A string
	B struct {
		X []int
	}
	C float64
}

var keyValuePairsInBlocks = []TestTableStruct{
	// Array inside struct with space separating elements followed by KV pair
	{contents: `A = some string 
				B { X = [1 2 3] }
				C = 12.34`, target: &ArrayInsideStructFollowedByKeyValPair{}, validateFunc: func(t interface{}) bool {
		v, ok := t.(*ArrayInsideStructFollowedByKeyValPair)
		return ok && v.B.X[0] == 1 && v.B.X[1] == 2 && v.B.X[2] == 3 && v.A == "some string" && v.C == 12.34
	}},
	// Array inside struct with space separating elements
	{contents: `A = some string 
				B { X = [1 2 3] }`, target: &ArrayInsideStruct{}, validateFunc: func(t interface{}) bool {
		v, ok := t.(*ArrayInsideStruct)
		return ok && v.B.X[0] == 1 && v.B.X[1] == 2 && v.B.X[2] == 3 && v.A == "some string"
	}},
	// Array inside struct with space separating elements
	{contents: `A = some string 
				B { X = [1 2 3] }`, target: &ArrayInsideStruct{}, validateFunc: func(t interface{}) bool {
		v, ok := t.(*ArrayInsideStruct)
		return ok && v.B.X[0] == 1 && v.B.X[1] == 2 && v.B.X[2] == 3 && v.A == "some string"
	}},
	// Array with space separating elements
	{contents: `X = [1 2 3]`, target: &IntArrayStruct{}, validateFunc: func(t interface{}) bool {
		v, ok := t.(*IntArrayStruct)
		return ok && v.X[0] == 1 && v.X[1] == 2 && v.X[2] == 3
	}},
	// array with commas separating elements
	{contents: `X = [1,2,3,4,5]`, target: &IntArrayStruct{}, validateFunc: func(t interface{}) bool {
		v, ok := t.(*IntArrayStruct)
		return ok && v.X[0] == 1 && v.X[1] == 2 && v.X[2] == 3
	}},
	{contents: `X = 10
	FloatValues {
		X = 10.857
	}
	StringValues {
		X = un quoted string
	}
	DurationValues {
		X = 10 seconds
	}`, target: &structWithSingleInnerBlock{}, validateFunc: func(t interface{}) bool {
		v, ok := t.(*structWithSingleInnerBlock)
		return ok && v.X == 10 && v.FloatValues.X == 10.857 && v.StringValues.X == "un quoted string" && v.DurationValues.X == 10*time.Second
	}},
	{contents: `
	Cluster {
		Name = "axlrate-charging"
		Connection {
			Host = "1.2.3.4"
			Port = 10080
			Timeout = 10 seconds
		}
	}
	`, target: &SysConfig{}, validateFunc: func(t interface{}) bool {
		v, ok := t.(*SysConfig)
		return ok && v.Cluster.Name == "axlrate-charging" && v.Cluster.Connection.Host == "1.2.3.4"
	}},
}

func TestKeyValuePairsInBlocks(t *testing.T) {
	for _, testcase := range keyValuePairsInBlocks {
		t.Log("Before : ", testcase.target)
		//t.Logf("input: %v", testcase.contents)
		parser := &HoconParser{}
		reader := strings.NewReader(testcase.contents)
		if err := parser.Parse(reader, testcase.target); err != nil {
			t.Errorf("failed for input : %v. Error : %v", testcase.contents, err)
		}
		if !testcase.validateFunc(testcase.target) {
			t.Errorf("input: %v", testcase.contents)
		}
		t.Log("After : ", testcase.target)
	}
}

func TestSingleLevelArrayOfObjects(t *testing.T) {
	fileContents := `
						X = [ 
							{ 
								A = "10", 
								B = 20 
							}, 
							{ 
								A = "30", 
								B = 40 
							} 
						]
						Y = 12.34
					`
	type TargetStruct struct {
		Y float64
		X []struct {
			A string
			B int
		}
	}
	target := &TargetStruct{}
	parser := &HoconParser{}
	reader := strings.NewReader(fileContents)
	if err := parser.Parse(reader, target); err != nil {
		t.Errorf("failed for input : %v. Error : %v", fileContents, err)
	}
	// Validate
	want := &TargetStruct{Y: 12.34, X: []struct {
		A string
		B int
	}{{"10", 20}, {"30", 40}}}
	if !(target.Y == 12.34 && target.X[0].A == "10" && target.X[1].A == "30") {
		t.Errorf("Got: %v, Want : %v", target, want)
	}
}

func TestUnBalancedParentheses(t *testing.T) {
	for _, test := range unbalancedParenTests {
		parser := &HoconParser{}
		reader := strings.NewReader(test.contents)
		if err := parser.Parse(reader, nil); !errorsAreEqual(err, test.err) {
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
