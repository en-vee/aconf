package aconf

import "fmt"

type LexLocation struct {
	lineNumber   int
	columnNumber int
}

type LexInvalidTokenErr struct {
	tokenValue string
	LexLocation
}

func (e *LexInvalidTokenErr) Error() string {
	return fmt.Sprintf("Invalid Token %s at location %d:%d", e.tokenValue, e.lineNumber, e.columnNumber)
}
