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

type ParserInvalidTargetErr struct{ got, want string }

func (err *ParserInvalidTargetErr) Error() string {
	return fmt.Sprintf("Invalid argument to method Parse. Want : %v, Got : %v", err.want, err.got)
}

type ParserInvalidTokenTypeErr struct {
	token HoconToken
}

func (err *ParserInvalidTokenTypeErr) Error() string {
	return fmt.Sprintf("parser: Invalid Token : %v", err.token)
}

type ParserInvalidInputFieldErr struct {
	fldName string
}

func (err *ParserInvalidInputFieldErr) Error() string {
	return fmt.Sprintf("parser: Invalid Field in Input : %v", err.fldName)
}
