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
	return fmt.Sprintf("lexer: %d:%d : Invalid Token %s", e.lineNumber, e.columnNumber, e.tokenValue)
}

type LexScannerErr struct {
	msg string
	LexLocation
}

func (e *LexScannerErr) Error() string {
	return fmt.Sprintf("lexer: %d:%d : %s", e.lineNumber, e.columnNumber, e.msg)
}
