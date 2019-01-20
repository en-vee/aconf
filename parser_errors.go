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
