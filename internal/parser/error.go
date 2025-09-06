package parser

import (
	"internal/scanner"
	"internal/util/log"
)

type ParseError struct {
	Message string
}

func NewParseErrorWithLog(message string, token *scanner.Token) *ParseError {
	err := &ParseError{
		Message: message,
	}

	log.Error("Parse error", log.E(err), log.A("token", token))

	return err
}

func (e *ParseError) Error() string {
	return e.Message
}
