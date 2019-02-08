package aconf

import "fmt"

type ParserUnbalancedParenthesesErr struct {
	LexLocation
}

func (err *ParserUnbalancedParenthesesErr) Error() string {
	return fmt.Sprintf("Unbalanced Parentheses")
}

type ParserInvalidRuneErr struct {
}

func (err *ParserInvalidRuneErr) Error() string {
	return fmt.Sprintf("Invalid Rune")
}

type ParserInvalidArrayErr struct {
	tokenValue string
	LexLocation
}

func (err *ParserInvalidArrayErr) Error() string {
	return fmt.Sprintf("parser: %d:%d : invalid array element %s", err.lineNumber, err.columnNumber, err.tokenValue)
}
