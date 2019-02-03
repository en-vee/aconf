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

type ErrReaderNil struct{}

func (e *ErrReaderNil) Error() string {
	return "lexer: io.Reader is nil"
}

type ErrLexerNotInitialized struct{}

func (l *ErrLexerNotInitialized) Error() string {
	return "Lexer Not created using aconf.NewLexer"
}

type ErrLexerInvalidToken struct {
	TokenText string
	LexLocation
}

func (e *ErrLexerInvalidToken) Error() string {
	return fmt.Sprintf("lexer: %d:%d Invalid Token %v", e.lineNumber, e.columnNumber, e.TokenText)
}

type ErrLexerInvalidDuration struct {
	val int
}

func (e *ErrLexerInvalidDuration) Error() string {
	return fmt.Sprintf("lexer: Invalid Duration : %d. Expecting +ve value", e.val)
}

type ErrLexerInvalidSize struct {
	val int
}

func (e *ErrLexerInvalidSize) Error() string {
	return fmt.Sprintf("lexer: Invalid Size : %d. Expecting +ve value", e.val)
}
